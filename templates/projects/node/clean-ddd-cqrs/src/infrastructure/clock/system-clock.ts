import { Clock } from '../../application/common/clock.js';

export class SystemClock extends Clock {
  now(): Date {
    return new Date();
  }
}
