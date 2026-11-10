use std::collections::HashMap;
use std::sync::{Arc, Mutex};

use async_trait::async_trait;
use chrono::{DateTime, TimeZone, Utc};
use sample_app::application::error::ApplicationError;
use sample_app::application::orders::{
    AddItem, EventPublisher, ListOrders, OrderRepository, OrderService,
};
use sample_app::domain::orders::{Order, OrderEvent};
use sample_app::domain::shared::Id;

#[derive(Default)]
struct FakeRepository {
    saved: Mutex<HashMap<Id, Order>>,
    limits: Mutex<Vec<usize>>,
}

#[async_trait]
impl OrderRepository for FakeRepository {
    async fn save(&self, order: &Order) -> Result<(), ApplicationError> {
        self.saved.lock().unwrap().insert(order.id(), order.clone());
        Ok(())
    }

    async fn find(&self, id: Id) -> Result<Option<Order>, ApplicationError> {
        Ok(self.saved.lock().unwrap().get(&id).cloned())
    }

    async fn list(&self, limit: usize) -> Result<Vec<Order>, ApplicationError> {
        self.limits.lock().unwrap().push(limit);
        Ok(Vec::new())
    }
}

#[derive(Default)]
struct FakeEvents {
    published: Mutex<Vec<OrderEvent>>,
}

impl EventPublisher for FakeEvents {
    fn publish(&self, events: &[OrderEvent]) {
        self.published.lock().unwrap().extend_from_slice(events);
    }
}

fn fixed_now() -> DateTime<Utc> {
    Utc.with_ymd_and_hms(2026, 1, 1, 0, 0, 0).unwrap()
}

fn service() -> (OrderService, Arc<FakeRepository>, Arc<FakeEvents>) {
    let repository = Arc::new(FakeRepository::default());
    let events = Arc::new(FakeEvents::default());
    let service = OrderService::new(repository.clone(), events.clone(), fixed_now);
    (service, repository, events)
}

fn pen(quantity: i32) -> AddItem {
    AddItem {
        product_name: "Caneta".to_owned(),
        quantity,
        unit_price: 2.5,
        currency: None,
    }
}

#[tokio::test]
async fn placing_publishes_event() {
    let (service, _, events) = service();
    let id = service.create("Ana").await.unwrap();
    service.add_item(id, pen(3)).await.unwrap();
    service.place(id).await.unwrap();

    let view = service.get(id).await.unwrap();
    assert_eq!(view.status, "Placed");
    assert_eq!(view.total, 7.5);
    assert_eq!(view.currency, "BRL");
    assert_eq!(view.created_at, fixed_now());
    assert_eq!(events.published.lock().unwrap().len(), 1);
}

#[tokio::test]
async fn rule_violation_is_reported_and_nothing_is_published() {
    let (service, _, events) = service();
    let id = service.create("Ana").await.unwrap();
    let error = service.add_item(id, pen(0)).await.unwrap_err();
    assert!(matches!(error, ApplicationError::Rule(_)));
    assert!(events.published.lock().unwrap().is_empty());
}

#[tokio::test]
async fn unknown_order_is_not_found() {
    let (service, _, _) = service();
    let error = service.place(Id::generate()).await.unwrap_err();
    assert!(matches!(error, ApplicationError::NotFound(_)));
}

#[tokio::test]
async fn list_limit_defaults_to_fifty_and_caps_at_two_hundred() {
    let (service, repository, _) = service();
    service.list(ListOrders::default()).await.unwrap();
    service
        .list(ListOrders { limit: Some(1000) })
        .await
        .unwrap();
    service.list(ListOrders { limit: Some(10) }).await.unwrap();
    assert_eq!(*repository.limits.lock().unwrap(), vec![50, 200, 10]);
}
