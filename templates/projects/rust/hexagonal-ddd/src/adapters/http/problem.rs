use axum::Json;
use axum::extract::rejection::{JsonRejection, QueryRejection};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Response};
use serde::Serialize;

use crate::application::error::ApplicationError;
use crate::domain::shared::RuleViolation;

#[derive(Debug, Serialize)]
pub struct Problem {
    status: u16,
    title: &'static str,
    detail: String,
}

impl Problem {
    pub fn new(status: StatusCode, title: &'static str, detail: impl Into<String>) -> Self {
        Self {
            status: status.as_u16(),
            title,
            detail: detail.into(),
        }
    }
}

impl IntoResponse for Problem {
    fn into_response(self) -> Response {
        let status = StatusCode::from_u16(self.status).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
        (status, Json(self)).into_response()
    }
}

impl From<ApplicationError> for Problem {
    fn from(error: ApplicationError) -> Self {
        match error {
            ApplicationError::NotFound(message) => {
                Self::new(StatusCode::NOT_FOUND, "Não encontrado", message)
            }
            ApplicationError::Rule(violation) => Self::from(violation),
            ApplicationError::Unavailable(message) => {
                tracing::error!(error = %message, "request failed");
                Self::new(
                    StatusCode::INTERNAL_SERVER_ERROR,
                    "Erro interno",
                    "Não foi possível concluir a operação.",
                )
            }
        }
    }
}

impl From<RuleViolation> for Problem {
    fn from(violation: RuleViolation) -> Self {
        Self::new(
            StatusCode::UNPROCESSABLE_ENTITY,
            "Regra de negócio violada",
            violation.to_string(),
        )
    }
}

impl From<JsonRejection> for Problem {
    fn from(rejection: JsonRejection) -> Self {
        Self::new(
            StatusCode::BAD_REQUEST,
            "Requisição inválida",
            rejection.body_text(),
        )
    }
}

impl From<QueryRejection> for Problem {
    fn from(rejection: QueryRejection) -> Self {
        Self::new(
            StatusCode::BAD_REQUEST,
            "Requisição inválida",
            rejection.body_text(),
        )
    }
}
