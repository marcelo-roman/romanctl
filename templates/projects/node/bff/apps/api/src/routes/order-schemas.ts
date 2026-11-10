import { CUSTOMER_NAME_MAX_LENGTH } from '../domain/order.js';
import { PRODUCT_NAME_MAX_LENGTH } from '../domain/order-item.js';

const item = {
  type: 'object',
  required: ['id', 'productName', 'quantity', 'unitPrice', 'subtotal'],
  properties: {
    id: { type: 'string' },
    productName: { type: 'string' },
    quantity: { type: 'integer' },
    unitPrice: { type: 'number' },
    subtotal: { type: 'number' },
  },
} as const;

const order = {
  type: 'object',
  required: ['id', 'customerName', 'status', 'createdAt', 'currency', 'total', 'canAddItems', 'canPlace', 'items'],
  properties: {
    id: { type: 'string' },
    customerName: { type: 'string' },
    status: { type: 'string' },
    createdAt: { type: 'string', format: 'date-time' },
    currency: { type: 'string' },
    total: { type: 'number' },
    canAddItems: { type: 'boolean' },
    canPlace: { type: 'boolean' },
    items: { type: 'array', items: item },
  },
} as const;

const idParams = {
  type: 'object',
  required: ['id'],
  properties: { id: { type: 'string', format: 'uuid' } },
} as const;

export const createOrderSchema = {
  body: {
    type: 'object',
    required: ['customerName'],
    additionalProperties: false,
    properties: { customerName: { type: 'string', maxLength: CUSTOMER_NAME_MAX_LENGTH } },
  },
  response: { 201: order },
} as const;

export const listOrdersSchema = {
  querystring: {
    type: 'object',
    additionalProperties: false,
    properties: { limit: { type: 'integer' } },
  },
  response: { 200: { type: 'array', items: order } },
} as const;

export const getOrderSchema = {
  params: idParams,
  response: { 200: order },
} as const;

export const addOrderItemSchema = {
  params: idParams,
  body: {
    type: 'object',
    required: ['productName', 'quantity', 'unitPrice'],
    additionalProperties: false,
    properties: {
      productName: { type: 'string', maxLength: PRODUCT_NAME_MAX_LENGTH },
      quantity: { type: 'number' },
      unitPrice: { type: 'number' },
      currency: { type: 'string' },
    },
  },
  response: { 200: order },
} as const;

export const placeOrderSchema = {
  params: idParams,
  response: { 200: order },
} as const;
