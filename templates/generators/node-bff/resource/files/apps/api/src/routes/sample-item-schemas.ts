import { SAMPLE_ITEM_NAME_MAX_LENGTH } from '../domain/sample-item.js';

const sampleItem = {
  type: 'object',
  required: ['id', 'name', 'createdAt'],
  properties: {
    id: { type: 'string' },
    name: { type: 'string' },
    createdAt: { type: 'string', format: 'date-time' },
  },
} as const;

const body = {
  type: 'object',
  required: ['name'],
  additionalProperties: false,
  properties: { name: { type: 'string', maxLength: SAMPLE_ITEM_NAME_MAX_LENGTH } },
} as const;

const idParams = {
  type: 'object',
  required: ['id'],
  properties: { id: { type: 'string', format: 'uuid' } },
} as const;

export const createSampleItemSchema = {
  body,
  response: { 201: sampleItem },
} as const;

export const listSampleItemsSchema = {
  querystring: {
    type: 'object',
    additionalProperties: false,
    properties: { limit: { type: 'integer' } },
  },
  response: { 200: { type: 'array', items: sampleItem } },
} as const;

export const getSampleItemSchema = {
  params: idParams,
  response: { 200: sampleItem },
} as const;

export const renameSampleItemSchema = {
  params: idParams,
  body,
  response: { 200: sampleItem },
} as const;

export const deleteSampleItemSchema = {
  params: idParams,
} as const;
