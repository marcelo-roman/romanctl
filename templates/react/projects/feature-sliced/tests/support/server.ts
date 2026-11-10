import { setupServer } from 'msw/node';
import { ordersApi } from './orders-api';

export const server = setupServer(...ordersApi);
