use chrono::{DateTime, TimeZone, Utc};
use sample_app::domain::orders::{Money, Order, OrderEvent, OrderStatus};

fn now() -> DateTime<Utc> {
    Utc.with_ymd_and_hms(2026, 1, 1, 12, 0, 0).unwrap()
}

fn brl(cents: i64) -> Money {
    Money::new(cents, "brl").unwrap()
}

#[test]
fn create_requires_customer() {
    assert!(Order::create("  ", now()).is_err());
    let order = Order::create(" Ana ", now()).unwrap();
    assert_eq!(order.customer_name(), "Ana");
    assert_eq!(order.status(), OrderStatus::Draft);
}

#[test]
fn total_sums_items() {
    let mut order = Order::create("Ana", now()).unwrap();
    order.add_item("Caneta", 2, brl(350)).unwrap();
    order.add_item("Caderno", 1, brl(1200)).unwrap();
    let total = order.total();
    assert_eq!(total.cents(), 1900);
    assert_eq!(total.currency(), "BRL");
}

#[test]
fn empty_order_totals_zero_in_default_currency() {
    let order = Order::create("Ana", now()).unwrap();
    assert_eq!(order.total(), Money::new(0, "BRL").unwrap());
}

#[test]
fn items_must_share_currency() {
    let mut order = Order::create("Ana", now()).unwrap();
    order.add_item("Caneta", 1, brl(100)).unwrap();
    let usd = Money::new(100, "USD").unwrap();
    let error = order.add_item("Pen", 1, usd).unwrap_err();
    assert_eq!(
        error.to_string(),
        "Todos os itens do pedido devem usar a mesma moeda."
    );
}

#[test]
fn item_requires_product_and_positive_quantity() {
    let mut order = Order::create("Ana", now()).unwrap();
    assert!(order.add_item(" ", 1, brl(100)).is_err());
    assert!(order.add_item("Caneta", 0, brl(100)).is_err());
    assert!(order.items().is_empty());
}

#[test]
fn place_raises_event_once() {
    let mut order = Order::create("Ana", now()).unwrap();
    assert!(order.place(now()).is_err());
    order.add_item("Caneta", 1, brl(100)).unwrap();
    order.place(now()).unwrap();

    let events = order.pull_events();
    assert_eq!(events.len(), 1);
    let OrderEvent::Placed(placed) = &events[0];
    assert_eq!(placed.order_id, order.id());
    assert_eq!(placed.total.cents(), 100);
    assert_eq!(events[0].name(), "order.placed");
    assert!(order.pull_events().is_empty());

    assert!(order.place(now()).is_err());
    assert!(order.add_item("Lápis", 1, brl(100)).is_err());
}

#[test]
fn money_rules() {
    assert!(Money::new(-1, "BRL").is_err());
    assert!(Money::new(100, "REAL").is_err());
    assert_eq!(Money::from_decimal(3.5, "BRL").unwrap().cents(), 350);
    assert!(brl(100).add(&Money::new(100, "USD").unwrap()).is_err());
    assert_eq!(brl(100).add(&brl(50)).unwrap().cents(), 150);
}
