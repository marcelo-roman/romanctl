import { CATALOG_NAME_MAX_LENGTH } from '../domain/catalog-product.js';

const product = {
  type: 'object',
  required: ['id', 'name', 'price', 'currency'],
  properties: {
    id: { type: 'string' },
    name: { type: 'string' },
    price: { type: 'number' },
    currency: { type: 'string' },
  },
} as const;

export const addCatalogProductSchema = {
  body: {
    type: 'object',
    required: ['name', 'price'],
    additionalProperties: false,
    properties: {
      name: { type: 'string', maxLength: CATALOG_NAME_MAX_LENGTH },
      price: { type: 'number' },
      currency: { type: 'string' },
    },
  },
  response: { 201: product },
} as const;

export const listCatalogProductsSchema = {
  response: { 200: { type: 'array', items: product } },
} as const;

export const getCatalogProductSchema = {
  params: {
    type: 'object',
    required: ['id'],
    properties: { id: { type: 'string', format: 'uuid' } },
  },
  response: { 200: product },
} as const;
