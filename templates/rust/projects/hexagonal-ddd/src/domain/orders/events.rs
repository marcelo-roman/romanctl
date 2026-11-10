use chrono::{DateTime, Utc};

use super::Money;
use crate::domain::shared::Id;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct OrderPlaced {
    pub order_id: Id,
    pub total: Money,
    pub occurred_at: DateTime<Utc>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum OrderEvent {
    Placed(OrderPlaced),
}

impl OrderEvent {
    pub fn name(&self) -> &'static str {
        match self {
            Self::Placed(_) => "order.placed",
        }
    }
}
