import type { FastifyError } from 'fastify';
import { ApiError, ApiUnavailableError } from '../api/api-error.js';
import { validationMessage } from './validation-message.js';

export interface UiError {
  status: number;
  message: string;
}

const UNAVAILABLE = 'The orders service did not answer. Try again in a moment.';
const UNEXPECTED = 'Something went wrong. Try again.';
const CHECK_DATA = 'Check the data you entered';

function fromApi(error: ApiError): UiError {
  if (error.status === 400) {
    return { status: 400, message: `${CHECK_DATA}: ${error.detail}` };
  }
  if (error.status === 404 || error.status === 422) {
    return { status: error.status, message: error.detail };
  }
  return { status: 502, message: UNEXPECTED };
}

export function uiError(error: FastifyError | Error): UiError {
  if (error instanceof ApiError) {
    return fromApi(error);
  }
  if (error instanceof ApiUnavailableError) {
    return { status: 502, message: UNAVAILABLE };
  }
  const request = error as FastifyError;
  if (request.validation) {
    return { status: 400, message: `${CHECK_DATA}: ${validationMessage(request.validation)}` };
  }
  if (request.statusCode && request.statusCode >= 400 && request.statusCode < 500) {
    return { status: request.statusCode, message: CHECK_DATA + '.' };
  }
  return { status: 500, message: UNEXPECTED };
}
