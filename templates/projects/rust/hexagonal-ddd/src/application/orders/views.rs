use chrono::{DateTime, Utc};

use crate::domain::orders::{Order, OrderItem};

#[derive(Debug, Clone, PartialEq)]
pub struct OrderItemView {
    pub product_name: String,
    pub quantity: i32,
    pub unit_price: f64,
    pub subtotal: f64,
}

#[derive(Debug, Clone, PartialEq)]
pub struct OrderView {
    pub id: String,
    pub customer_name: String,
    pub status: String,
    pub created_at: DateTime<Utc>,
    pub total: f64,
    pub currency: String,
    pub items: Vec<OrderItemView>,
}

impl From<&OrderItem> for OrderItemView {
    fn from(item: &OrderItem) -> Self {
        Self {
            product_name: item.product_name().to_owned(),
            quantity: item.quantity(),
            unit_price: item.unit_price().amount(),
            subtotal: item.subtotal().amount(),
        }
    }
}

impl From<&Order> for OrderView {
    fn from(order: &Order) -> Self {
        let total = order.total();
        Self {
            id: order.id().to_string(),
            customer_name: order.customer_name().to_owned(),
            status: order.status().as_str().to_owned(),
            created_at: order.created_at(),
            total: total.amount(),
            currency: total.currency().to_owned(),
            items: order.items().iter().map(OrderItemView::from).collect(),
        }
    }
}
