const { randomUUID } = require('node:crypto');
const express = require('express');

const router = express.Router();
const items = new Map();

router.get('/', (req, res) => res.json([...items.values()]));

router.get('/:id', (req, res) => {
  const item = items.get(req.params.id);
  return item ? res.json(item) : res.sendStatus(404);
});

router.post('/', (req, res) => {
  const item = { ...req.body, id: randomUUID() };
  items.set(item.id, item);
  res.status(201).json(item);
});

router.delete('/:id', (req, res) => res.sendStatus(items.delete(req.params.id) ? 204 : 404));

module.exports = router;
