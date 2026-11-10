use std::sync::Arc;

use super::{AddItem, EventPublisher, ListOrders, OrderRepository, OrderView};
use crate::application::clock::Clock;
use crate::application::error::ApplicationError;
use crate::domain::orders::{DEFAULT_CURRENCY, Money, Order};
use crate::domain::shared::{Id, RuleViolation};

pub struct OrderService {
    repository: Arc<dyn OrderRepository>,
    events: Arc<dyn EventPublisher>,
    clock: Clock,
}

impl OrderService {
    pub fn new(
        repository: Arc<dyn OrderRepository>,
        events: Arc<dyn EventPublisher>,
        clock: Clock,
    ) -> Self {
        Self {
            repository,
            events,
            clock,
        }
    }

    pub async fn create(&self, customer_name: &str) -> Result<Id, ApplicationError> {
        let order = Order::create(customer_name, (self.clock)())?;
        self.repository.save(&order).await?;
        Ok(order.id())
    }

    pub async fn add_item(&self, id: Id, command: AddItem) -> Result<(), ApplicationError> {
        let currency = command
            .currency
            .as_deref()
            .filter(|currency| !currency.trim().is_empty())
            .unwrap_or(DEFAULT_CURRENCY);
        let unit_price = Money::from_decimal(command.unit_price, currency)?;
        self.change(id, |order| {
            order.add_item(&command.product_name, command.quantity, unit_price)
        })
        .await
    }

    pub async fn place(&self, id: Id) -> Result<(), ApplicationError> {
        let now = (self.clock)();
        self.change(id, |order| order.place(now)).await
    }

    pub async fn get(&self, id: Id) -> Result<OrderView, ApplicationError> {
        let order = self.load(id).await?;
        Ok(OrderView::from(&order))
    }

    pub async fn list(&self, query: ListOrders) -> Result<Vec<OrderView>, ApplicationError> {
        let orders = self.repository.list(query.limit()).await?;
        Ok(orders.iter().map(OrderView::from).collect())
    }

    async fn change(
        &self,
        id: Id,
        apply: impl FnOnce(&mut Order) -> Result<(), RuleViolation>,
    ) -> Result<(), ApplicationError> {
        let mut order = self.load(id).await?;
        apply(&mut order)?;
        let events = order.pull_events();
        self.repository.save(&order).await?;
        self.events.publish(&events);
        Ok(())
    }

    async fn load(&self, id: Id) -> Result<Order, ApplicationError> {
        self.repository
            .find(id)
            .await?
            .ok_or_else(|| ApplicationError::not_found("Pedido não encontrado."))
    }
}
