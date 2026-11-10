use std::fmt::Display;

use crate::domain::shared::RuleViolation;

#[derive(Debug, thiserror::Error)]
pub enum ApplicationError {
    #[error("{0}")]
    NotFound(String),
    #[error(transparent)]
    Rule(#[from] RuleViolation),
    #[error("{0}")]
    Unavailable(String),
}

impl ApplicationError {
    pub fn not_found(message: impl Into<String>) -> Self {
        Self::NotFound(message.into())
    }

    pub fn unavailable(error: impl Display) -> Self {
        Self::Unavailable(error.to_string())
    }
}
