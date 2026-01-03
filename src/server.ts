import express from 'express';
import { IncomingMessage, ServerResponse, ClientRequest } from 'http';
import { createProxyMiddleware, responseInterceptor, Options } from 'http-proxy-middleware';
import cors from 'cors';
import { v4 as uuidv4 } from 'uuid';
import promClient from 'prom-client';
import path from 'path';

import { config } from './config';
import { redisService } from './services/redis';
import { chaosEngine } from './chaos';
import { configController } from './controllers/configController';
import { authMiddleware } from './middleware/auth';
import { ChaosConfig, RequestLog } from './models/types';

// Metrics Setup
const collectDefaultMetrics = promClient.collectDefaultMetrics;
collectDefaultMetrics();

const requestCounter = new promClient.Counter({
  name: 'chaos_proxy_requests_total',
  help: 'Total requests',
  labelNames: ['config_id', 'status_code', 'chaos_type']
});

const latencyHistogram = new promClient.Histogram({
  name: 'chaos_proxy_request_duration_seconds',
  help: 'Request latency',
  labelNames: ['config_id', 'chaos_type']
});

const app = express();

// Basic Middleware
app.use(cors());
app.use(express.json()); // NOTE: interfere with proxy body? Only for API routes. 
// We should apply json parser only to API routes to avoid messing with proxy streams.

// Static UI
app.use('/dashboard', express.static(path.join(__dirname, '../web')));

// API Routes
const apiRouter = express.Router();
apiRouter.use(express.json());
apiRouter.use(authMiddleware);
apiRouter.post('/configs', configController.create);
apiRouter.get('/configs', configController.list);
apiRouter.get('/configs/:id', configController.get);
apiRouter.put('/configs/:id', configController.update);
apiRouter.delete('/configs/:id', configController.delete);
apiRouter.get('/logs', configController.getLogs);

app.use('/api/v1', apiRouter);
// Alias
app.use('/rules', apiRouter);

// Metrics
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', promClient.register.contentType);
  res.end(await promClient.register.metrics());
});

// PROXY LOGIC
// We capture specific path or header
app.use(async (req, res, next) => {
    // Skip API, Metrics, Dashboard
    if (req.path.startsWith('/api/') || req.path.startsWith('/dashboard') || req.path.startsWith('/metrics') || req.path.startsWith('/rules')) {
        return next();
    }

    const start = Date.now();
    let configId = req.headers['x-chaos-config-id'] as string;
    
    // Path based: /proxy/:id/foo
    if (req.path.startsWith('/proxy/')) {
        const parts = req.path.split('/');
        if (parts.length >= 3) {
            configId = parts[2];
            // Rewrite path for later? handled by proxy rewrite
        }
    }

    if (!configId) {
        // Not a chaos request (or root) -> 404
        return res.status(404).send('Chaos Proxy: Missing Config ID');
    }

    // Load Config
    try {
        const cfg = await redisService.getConfig(configId);
        if (!cfg || !cfg.enabled) {
            return res.status(404).send('Chaos Config not found or disabled');
        }

        // Make Chaos Decision
        const decision = chaosEngine.decide(cfg.rules, req);
        const chaosType = decision.shouldError ? 'error' : (decision.shouldLatency ? 'latency' : 'none');

        // Store for logging in onFinish
        const logData: RequestLog = {
            id: uuidv4(),
            timestamp: new Date().toISOString(),
            config_id: configId,
            method: req.method,
            path: req.originalUrl,
            status_code: 200,
            duration_ms: 0,
            chaos_type: chaosType
        };

        // 1. Error Injection
        if (decision.shouldError) {
            for (const [k, v] of Object.entries(decision.headers)) res.set(k, v);
            res.status(decision.errorCode).send(decision.errorBody);
            
            // Log
            logData.status_code = decision.errorCode;
            logData.duration_ms = Date.now() - start;
            requestCounter.inc({ config_id: configId, status_code: decision.errorCode, chaos_type: 'error' });
            redisService.logRequest(logData);
            return;
        }

        // 2. Latency Injection
        if (decision.shouldLatency) {
            await new Promise(r => setTimeout(r, decision.latencyMs));
        }

        // Attach data to req for proxy middleware to use
        (req as any).chaosConfig = cfg;
        (req as any).chaosDecision = decision;
        (req as any).chaosStartTime = start;
        (req as any).chaosLogData = logData;

        next(); // Proceed to proxy middleware

    } catch (e) {
        console.error(e);
        res.status(500).send('Internal Proxy Error');
    }
});

// Proxy Middleware
const proxy = createProxyMiddleware({
    router: (req) => {
        const cfg = (req as any).chaosConfig as ChaosConfig;
        return cfg.target; // Dynamic target
    },
    pathRewrite: (path, req) => {
        const cfg = (req as any).chaosConfig as ChaosConfig;
        if (path.startsWith(`/proxy/${cfg.id}`)) {
            return path.replace(`/proxy/${cfg.id}`, '') || '/';
        }
        return path;
    },
    changeOrigin: true,
    selfHandleResponse: true, 
    on: {
        proxyReq: (proxyReq: ClientRequest, req: IncomingMessage, res: ServerResponse) => {
            const decision = (req as any).chaosDecision;
            if (decision && decision.headers) {
                Object.entries(decision.headers).forEach(([k, v]) => {
                     proxyReq.setHeader(k, v as string);
                });
            }
            proxyReq.setHeader('X-Chaos-Proxy', 'true');
        },
        proxyRes: responseInterceptor(async (responseBuffer, proxyRes, req, res) => {
            const decision = (req as any).chaosDecision;
            const cfg = (req as any).chaosConfig as ChaosConfig;
            
            let buffer = responseBuffer;
    
            // Fuzzing
            if (decision.shouldFuzz && cfg.rules.response_fuzzing) {
                try {
                    const bodyStr = responseBuffer.toString('utf8');
                    const mutated = chaosEngine.fuzzBody(bodyStr, cfg.rules.response_fuzzing.mutation_rate || 0.1);
                    buffer = Buffer.from(JSON.stringify(mutated));
                    res.setHeader('X-Chaos-Proxy-Fuzzed', 'true');
                } catch (e) {
                    // Fuzz failed
                }
            }
    
            // Metrics & Logging
            const start = (req as any).chaosStartTime;
            const logData = (req as any).chaosLogData as RequestLog;
            const duration = Date.now() - start;
    
            requestCounter.inc({ 
                config_id: cfg.id, 
                status_code: res.statusCode, 
                chaos_type: decision.shouldFuzz ? 'fuzzing' : (decision.shouldLatency ? 'latency' : 'none') 
            });
            latencyHistogram.observe({ 
                config_id: cfg.id, 
                chaos_type: decision.shouldFuzz ? 'fuzzing' : (decision.shouldLatency ? 'latency' : 'none') 
            }, duration / 1000);
    
            logData.status_code = res.statusCode;
            logData.duration_ms = duration;
            redisService.logRequest(logData).catch(console.error);
    
            return buffer;
        }),
    }
});

// Use proxy for everything else
app.use('/', proxy);

app.listen(config.port, () => {
  console.log(`🌪️ Chaos Proxy (Titanium Edition/Node) running on port ${config.port}`);
});
