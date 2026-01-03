import Redis from 'ioredis';
import { config } from '../config';
import { ChaosConfig, RequestLog } from '../models/types';

const redis = new Redis({
  host: config.redis.host,
  port: config.redis.port,
  password: config.redis.password,
  db: config.redis.db,
  retryStrategy: (times) => Math.min(times * 50, 2000), // Retry with backoff
});

const KEY_PREFIX = 'chaos:config:';
const LIST_KEY = 'chaos:configs';
const LOGS_KEY = 'chaos:logs:global';
const MAX_LOGS = 100;

export const redisService = {
  client: redis,

  async saveConfig(cfg: ChaosConfig): Promise<void> {
    const key = KEY_PREFIX + cfg.id;
    await redis.multi()
      .set(key, JSON.stringify(cfg))
      .sadd(LIST_KEY, cfg.id)
      .exec();
  },

  async getConfig(id: string): Promise<ChaosConfig | null> {
    const data = await redis.get(KEY_PREFIX + id);
    return data ? JSON.parse(data) : null;
  },

  async listConfigs(): Promise<ChaosConfig[]> {
    const ids = await redis.smembers(LIST_KEY);
    if (ids.length === 0) return [];
    
    // Fetch all configs in parallel
    const keys = ids.map(id => KEY_PREFIX + id);
    const results = await redis.mget(...keys);
    
    return results
      .filter((r): r is string => r !== null)
      .map(r => JSON.parse(r) as ChaosConfig);
  },

  async deleteConfig(id: string): Promise<void> {
    await redis.multi()
      .del(KEY_PREFIX + id)
      .srem(LIST_KEY, id)
      .exec();
  },

  async logRequest(log: RequestLog): Promise<void> {
    const data = JSON.stringify(log);
    // Lua script implementation of LPUSH + LTRIM usually better for atomicity 
    // but pipeline is fine here.
    await redis.pipeline()
      .lpush(LOGS_KEY, data)
      .ltrim(LOGS_KEY, 0, MAX_LOGS - 1)
      .exec();
  },

  async getLogs(limit: number = 50): Promise<RequestLog[]> {
    const rawLogs = await redis.lrange(LOGS_KEY, 0, limit - 1);
    return rawLogs
      .map(l => {
        try { return JSON.parse(l); } 
        catch { return null; }
      })
      .filter(l => l !== null);
  }
};
