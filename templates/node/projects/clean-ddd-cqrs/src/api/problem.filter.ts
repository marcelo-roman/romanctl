import { type ArgumentsHost, Catch, type ExceptionFilter, HttpException, HttpStatus, Logger } from '@nestjs/common';
import { NotFoundError } from '../application/common/not-found-error.js';
import { DomainError } from '../domain/shared/domain-error.js';

interface Problem {
  status: number;
  title: string;
  detail?: string;
}

interface ProblemResponse {
  status(code: number): ProblemResponse;
  type(contentType: string): ProblemResponse;
  send(body: string): unknown;
}

const titles: Record<number, string> = {
  [HttpStatus.BAD_REQUEST]: 'Invalid request',
  [HttpStatus.NOT_FOUND]: 'Not found',
  [HttpStatus.UNPROCESSABLE_ENTITY]: 'Business rule violated',
  [HttpStatus.INTERNAL_SERVER_ERROR]: 'Internal error',
};

@Catch()
export class ProblemFilter implements ExceptionFilter {
  private readonly logger = new Logger(ProblemFilter.name);

  catch(exception: unknown, host: ArgumentsHost): void {
    const problem = this.problemOf(exception);
    host.switchToHttp().getResponse<ProblemResponse>().status(problem.status).type('application/problem+json').send(JSON.stringify(problem));
  }

  private problemOf(exception: unknown): Problem {
    if (exception instanceof DomainError) {
      return this.problem(HttpStatus.UNPROCESSABLE_ENTITY, exception.message);
    }
    if (exception instanceof NotFoundError) {
      return this.problem(HttpStatus.NOT_FOUND, exception.message);
    }
    if (exception instanceof HttpException) {
      return this.problem(exception.getStatus(), exception.message);
    }
    this.logger.error(exception);
    return { status: HttpStatus.INTERNAL_SERVER_ERROR, title: titles[HttpStatus.INTERNAL_SERVER_ERROR] ?? '' };
  }

  private problem(status: number, detail: string): Problem {
    return { status, title: titles[status] ?? detail, detail };
  }
}
