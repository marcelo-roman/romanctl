import { TITLE_MAX_LENGTH } from '../models/task.js';

const task = {
  type: 'object',
  required: ['id', 'title', 'done', 'dueDate', 'createdAt'],
  properties: {
    id: { type: 'string' },
    title: { type: 'string' },
    done: { type: 'boolean' },
    dueDate: { type: ['string', 'null'] },
    createdAt: { type: 'string', format: 'date-time' },
  },
} as const;

const title = { type: 'string', minLength: 1, maxLength: TITLE_MAX_LENGTH } as const;

const dueDate = { type: ['string', 'null'], format: 'date' } as const;

const idParams = {
  type: 'object',
  required: ['id'],
  properties: { id: { type: 'string', format: 'uuid' } },
} as const;

export const createTaskSchema = {
  body: {
    type: 'object',
    required: ['title'],
    additionalProperties: false,
    properties: { title, dueDate },
  },
  response: { 201: task },
} as const;

export const listTasksSchema = {
  querystring: {
    type: 'object',
    additionalProperties: false,
    properties: {
      done: { type: 'boolean' },
      limit: { type: 'integer' },
    },
  },
  response: { 200: { type: 'array', items: task } },
} as const;

export const getTaskSchema = {
  params: idParams,
  response: { 200: task },
} as const;

export const updateTaskSchema = {
  params: idParams,
  body: {
    type: 'object',
    additionalProperties: false,
    minProperties: 1,
    properties: { title, done: { type: 'boolean' }, dueDate },
  },
  response: { 200: task },
} as const;

export const deleteTaskSchema = {
  params: idParams,
} as const;
