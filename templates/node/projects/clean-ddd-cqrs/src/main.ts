import 'reflect-metadata';
import { Logger } from '@nestjs/common';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './api/app.module.js';
import { PostgresPersistenceModule } from './infrastructure/persistence/postgres/postgres-persistence.module.js';

const port = Number(process.env.PORT) || 3000;
const app = await NestFactory.create(AppModule.with(PostgresPersistenceModule));
app.enableShutdownHooks();
await app.listen(port);
new Logger('SampleApp').log(`http://localhost:${port}`);
