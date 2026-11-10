import { type DynamicModule, Module, type Type } from '@nestjs/common';
import { APP_FILTER } from '@nestjs/core';
import { CqrsModule } from '@nestjs/cqrs';
import { Clock } from '../application/common/clock.js';
import { orderHandlers } from '../application/orders/order-handlers.js';
import { SystemClock } from '../infrastructure/clock/system-clock.js';
import { HealthController } from './health.controller.js';
import { OrdersController } from './orders.controller.js';
import { ProblemFilter } from './problem.filter.js';
// romanctl:imports

@Module({})
export class AppModule {
  static with(persistence: Type): DynamicModule {
    return {
      module: AppModule,
      imports: [CqrsModule.forRoot(), persistence],
      controllers: [
        HealthController,
        OrdersController,
        // romanctl:controllers
      ],
      providers: [
        { provide: Clock, useClass: SystemClock },
        { provide: APP_FILTER, useClass: ProblemFilter },
        ...orderHandlers,
        // romanctl:handlers
      ],
    };
  }
}
