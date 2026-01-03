import vm from 'vm';
import { ChaosDecision } from './index';

export class ScriptEngine {
  /**
   * Executes a user-provided JS script to modify the chaos decision.
   * The script has access to 'req' (readonly) and 'decision' (mutable).
   */
  execute(script: string, context: { req: any; decision: ChaosDecision }): void {
    if (!script || script.trim() === '') return;

    try {
      // Create a sandbox
      // We purposefully limit what's exposed to avoid security risks
      const sandbox = { 
        req: {
            method: context.req.method,
            path: context.req.path || context.req.url, // Handle both express/node
            headers: context.req.headers,
            body: context.req.body,
            query: context.req.query || {}
        },
        decision: context.decision,
        
        // Allowed utils
        Math: Math,
        console: { log: (...args: any[]) => console.log('[SCRIPT]', ...args) }, // Allow logging with prefix
        Date: Date
      };
      
      vm.createContext(sandbox);
      
      // Execute script with strict limits
      // This is dynamic code execution! 
      vm.runInContext(script, sandbox, { 
          timeout: 50, // 50ms max execution time to prevent DoS
          displayErrors: false
      });
      
    } catch (e) {
      console.warn('JS Script execution failed:', e);
      // We suppress script errors so they don't crash the proxy, 
      // chaos simply won't be applied as intended by the script.
    }
  }
}

export const scriptEngine = new ScriptEngine();
