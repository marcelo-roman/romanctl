import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { server } from '@tests/support';
import { errorMessage } from './error-message';
import { getJson, postJson } from './http-client';

describe('http client', () => {
  it('sends JSON to the API under the base URL', async () => {
    server.use(http.post('*/api/echo', async ({ request }) => HttpResponse.json(await request.json(), { status: 201 })));

    await expect(postJson('/echo', { a: 1 })).resolves.toEqual({ a: 1 });
  });

  it('resolves a 204 to undefined', async () => {
    server.use(http.post('*/api/empty', () => new HttpResponse(null, { status: 204 })));

    await expect(postJson('/empty')).resolves.toBeUndefined();
  });

  it('turns a problem response into an ApiError with the detail', async () => {
    server.use(http.get('*/api/broken', () => HttpResponse.json({ status: 422, title: 'Rule', detail: 'Not allowed.' }, { status: 422 })));

    const error = await getJson('/broken').catch((caught: unknown) => caught);

    expect(error).toMatchObject({ status: 422, title: 'Rule', detail: 'Not allowed.' });
    expect(errorMessage(error)).toBe('Not allowed.');
  });

  it('falls back to a default title when the body is not a problem', async () => {
    server.use(http.get('*/api/plain', () => new HttpResponse('oops', { status: 404 })));

    const error = await getJson('/plain').catch((caught: unknown) => caught);

    expect(errorMessage(error)).toBe('Not found.');
  });

  it('explains when the API cannot be reached', async () => {
    server.use(http.get('*/api/down', () => HttpResponse.error()));

    const error = await getJson('/down').catch((caught: unknown) => caught);

    expect(errorMessage(error)).toBe('Could not reach the API. Check that it is running.');
  });
});
