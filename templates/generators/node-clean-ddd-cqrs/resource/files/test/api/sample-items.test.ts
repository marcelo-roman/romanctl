import type { INestApplication } from '@nestjs/common';
import request from 'supertest';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { startApp } from './app.js';

describe('SampleItems API', () => {
  let app: INestApplication;
  let http: ReturnType<typeof request>;

  beforeEach(async () => {
    app = await startApp();
    http = request(app.getHttpServer());
  });

  afterEach(() => app.close());

  it('creates, renames, reads and deletes', async () => {
    const created = await http.post('/sample-items').send({ name: 'Primeiro' }).expect(201);
    const id: string = created.body.id;
    expect(created.headers.location).toBe(`/sample-items/${id}`);

    await http.put(`/sample-items/${id}`).send({ name: 'Segundo' }).expect(204);
    const got = await http.get(`/sample-items/${id}`).expect(200);
    expect(got.body).toMatchObject({ id, name: 'Segundo' });

    const list = await http.get('/sample-items?limit=10').expect(200);
    expect(list.body.map((item: { id: string }) => item.id)).toEqual([id]);

    await http.delete(`/sample-items/${id}`).expect(204);
    await http.get(`/sample-items/${id}`).expect(404);
  });

  it('answers 422 for an empty name and 400 for an invalid body', async () => {
    await http.post('/sample-items').send({ name: ' ' }).expect(422);
    await http.post('/sample-items').send({ name: 1 }).expect(400);
  });
});
