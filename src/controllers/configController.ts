import { Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { redisService } from '../services/redis';
import { ChaosConfig } from '../models/types';

export const configController = {
  async create(req: Request, res: Response) {
    try {
      const cfg: ChaosConfig = req.body;
      cfg.id = cfg.id || uuidv4();
      cfg.created_at = new Date().toISOString();
      cfg.updated_at = new Date().toISOString();
      cfg.enabled = cfg.enabled !== undefined ? cfg.enabled : true;

      await redisService.saveConfig(cfg);
      res.status(201).json(cfg);
    } catch (err) {
      res.status(500).json({ error: 'Failed to save config' });
    }
  },

  async list(req: Request, res: Response) {
    try {
      const configs = await redisService.listConfigs();
      res.json({ configs, count: configs.length });
    } catch (err) {
      res.status(500).json({ error: 'Failed to list configs' });
    }
  },

  async get(req: Request, res: Response) {
    try {
      const cfg = await redisService.getConfig(req.params.id);
      if (!cfg) return res.status(404).json({ error: 'Not found' });
      res.json(cfg);
    } catch (err) {
      res.status(500).json({ error: 'Failed to get config' });
    }
  },

  async update(req: Request, res: Response) {
    try {
      const existing = await redisService.getConfig(req.params.id);
      if (!existing) return res.status(404).json({ error: 'Not found' });

      const cfg: ChaosConfig = { ...existing, ...req.body, id: existing.id, updated_at: new Date().toISOString() };
      await redisService.saveConfig(cfg);
      res.json(cfg);
    } catch (err) {
      res.status(500).json({ error: 'Failed to update config' });
    }
  },

  async delete(req: Request, res: Response) {
    try {
      await redisService.deleteConfig(req.params.id);
      res.status(200).json({ message: 'Deleted', id: req.params.id });
    } catch (err) {
      res.status(500).json({ error: 'Failed to delete config' });
    }
  },

  async getLogs(req: Request, res: Response) {
    try {
        const limit = parseInt(req.query.limit as string) || 50;
        const logs = await redisService.getLogs(limit);
        res.json({ logs });
    } catch (err) {
        res.status(500).json({ error: 'Failed to get logs' });
    }
  }
};
