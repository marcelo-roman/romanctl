use chrono::{DateTime, TimeZone, Utc};
use sample_app::domain::orders::{Money, Order, OrderEvent, OrderStatus};

fn now() -> DateTime<Utc> {
    Utc.with_ymd_and_hms(2026, 1, 1, 12, 0, 0).unwrap()
}

fn usd(cents: i64) -> Money {
    Money::new(cents, "usd").unwrap()
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
    order.add_item("Pen", 2, usd(350)).unwrap();
    order.add_item("Notebook", 1, usd(1200)).unwrap();
    let total = order.total();
    assert_eq!(total.cents(), 1900);
    assert_eq!(total.currency(), "USD");
}

#[test]
fn empty_order_totals_zero_in_default_currency() {
    let order = Order::create("Ana", now()).unwrap();
    assert_eq!(order.total(), Money::new(0, "USD").unwrap());
}

#[test]
fn items_must_share_currency() {
    let mut order = Order::create("Ana", now()).unwrap();
    order.add_item("Pen", 1, usd(100)).unwrap();
    let eur = Money::new(100, "EUR").unwrap();
    let error = order.add_item("Pen", 1, eur).unwrap_err();
    assert_eq!(
        error.to_string(),
        "All order items must use the same currency."
    );
}

#[test]
fn item_requires_product_and_positive_quantity() {
    let mut order = Order::create("Ana", now()).unwrap();
    assert!(order.add_item(" ", 1, usd(100)).is_err());
    assert!(order.add_item("Pen", 0, usd(100)).is_err());
    assert!(order.items().is_empty());
}

#[test]
fn place_raises_event_once() {
    let mut order = Order::create("Ana", now()).unwrap();
    assert!(order.place(now()).is_err());
    order.add_item("Pen", 1, usd(100)).unwrap();
    order.place(now()).unwrap();

    let events = order.pull_events();
    assert_eq!(events.len(), 1);
    let OrderEvent::Placed(placed) = &events[0];
    assert_eq!(placed.order_id, order.id());
    assert_eq!(placed.total.cents(), 100);
    assert_eq!(events[0].name(), "order.placed");
    assert!(order.pull_events().is_empty());

    assert!(order.place(now()).is_err());
    assert!(order.add_item("Pencil", 1, usd(100)).is_err());
}

#[test]
fn money_rules() {
    assert!(Money::new(-1, "USD").is_err());
    assert!(Money::new(100, "REAL").is_err());
    assert_eq!(Money::from_decimal(3.5, "USD").unwrap().cents(), 350);
    assert!(usd(100).add(&Money::new(100, "EUR").unwrap()).is_err());
    assert_eq!(usd(100).add(&usd(50)).unwrap().cents(), 150);
}
