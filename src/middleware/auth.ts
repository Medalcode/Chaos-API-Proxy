import { Request, Response, NextFunction } from 'express';
import { config } from '../config';

const keys = new Set(config.apiKeys);

export const authMiddleware = (req: Request, res: Response, next: NextFunction) => {
  if (keys.size === 0) return next(); // Auth disabled

  const apiKey = req.header('X-API-Key') || req.query.api_key?.toString();
  
  if (!apiKey || !keys.has(apiKey)) {
    return res.status(401).json({ error: 'Unauthorized: Invalid API Key' });
  }
  
  next();
};
