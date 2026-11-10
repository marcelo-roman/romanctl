use chrono::{DateTime, Utc};

use super::{DEFAULT_CURRENCY, Money, OrderEvent, OrderItem, OrderPlaced, OrderStatus};
use crate::domain::shared::{Id, RuleViolation};

#[derive(Debug, Clone)]
pub struct Order {
    id: Id,
    customer_name: String,
    status: OrderStatus,
    created_at: DateTime<Utc>,
    items: Vec<OrderItem>,
    events: Vec<OrderEvent>,
}

impl Order {
    pub fn create(customer_name: &str, now: DateTime<Utc>) -> Result<Self, RuleViolation> {
        let customer_name = customer_name.trim();
        if customer_name.is_empty() {
            return Err(RuleViolation::new("O nome do cliente é obrigatório."));
        }
        Ok(Self {
            id: Id::generate(),
            customer_name: customer_name.to_owned(),
            status: OrderStatus::Draft,
            created_at: now,
            items: Vec::new(),
            events: Vec::new(),
        })
    }

    pub fn restore(
        id: Id,
        customer_name: String,
        status: OrderStatus,
        created_at: DateTime<Utc>,
        items: Vec<OrderItem>,
    ) -> Self {
        Self {
            id,
            customer_name,
            status,
            created_at,
            items,
            events: Vec::new(),
        }
    }

    pub fn id(&self) -> Id {
        self.id
    }

    pub fn customer_name(&self) -> &str {
        &self.customer_name
    }

    pub fn status(&self) -> OrderStatus {
        self.status
    }

    pub fn created_at(&self) -> DateTime<Utc> {
        self.created_at
    }

    pub fn items(&self) -> &[OrderItem] {
        &self.items
    }

    pub fn currency(&self) -> &str {
        self.items
            .first()
            .map_or(DEFAULT_CURRENCY, |item| item.unit_price().currency())
    }

    pub fn total(&self) -> Money {
        let cents = self.items.iter().map(|item| item.subtotal().cents()).sum();
        Money::trusted(cents, self.currency())
    }

    pub fn add_item(
        &mut self,
        product_name: &str,
        quantity: i32,
        unit_price: Money,
    ) -> Result<(), RuleViolation> {
        if self.status != OrderStatus::Draft {
            return Err(RuleViolation::new(
                "Um pedido já realizado não pode ser alterado.",
            ));
        }
        if !self.items.is_empty() && unit_price.currency() != self.currency() {
            return Err(RuleViolation::new(
                "Todos os itens do pedido devem usar a mesma moeda.",
            ));
        }
        self.items
            .push(OrderItem::create(product_name, quantity, unit_price)?);
        Ok(())
    }

    pub fn place(&mut self, now: DateTime<Utc>) -> Result<(), RuleViolation> {
        if self.status != OrderStatus::Draft {
            return Err(RuleViolation::new("O pedido já foi realizado."));
        }
        if self.items.is_empty() {
            return Err(RuleViolation::new(
                "Um pedido sem itens não pode ser realizado.",
            ));
        }
        self.status = OrderStatus::Placed;
        self.events.push(OrderEvent::Placed(OrderPlaced {
            order_id: self.id,
            total: self.total(),
            occurred_at: now,
        }));
        Ok(())
    }

    pub fn pull_events(&mut self) -> Vec<OrderEvent> {
        std::mem::take(&mut self.events)
    }
}
