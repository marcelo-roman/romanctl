use std::collections::HashMap;

use async_trait::async_trait;
use chrono::{DateTime, Utc};
use sqlx::postgres::PgRow;
use sqlx::{Row, query};
use uuid::Uuid;

use super::Database;
use crate::application::error::ApplicationError;
use crate::application::orders::OrderRepository;
use crate::domain::orders::{Money, Order, OrderItem, OrderStatus};
use crate::domain::shared::Id;

pub struct PostgresOrderRepository {
    database: Database,
}

impl PostgresOrderRepository {
    pub fn new(database: Database) -> Self {
        Self { database }
    }

    async fn load(&self, rows: Vec<PgRow>) -> Result<Vec<Order>, ApplicationError> {
        let ids: Vec<Uuid> = rows
            .iter()
            .map(|row| row.try_get("id"))
            .collect::<Result<_, _>>()
            .map_err(ApplicationError::unavailable)?;
        let mut items = self.items(&ids).await?;
        rows.iter()
            .map(|row| {
                let id: Uuid = row.try_get("id").map_err(ApplicationError::unavailable)?;
                restore_order(row, items.remove(&id).unwrap_or_default())
            })
            .collect()
    }

    async fn items(
        &self,
        order_ids: &[Uuid],
    ) -> Result<HashMap<Uuid, Vec<OrderItem>>, ApplicationError> {
        let mut items: HashMap<Uuid, Vec<OrderItem>> = HashMap::new();
        if order_ids.is_empty() {
            return Ok(items);
        }
        let rows = query(
            "SELECT order_id, id, product_name, quantity, unit_price, currency
             FROM order_items WHERE order_id = ANY($1) ORDER BY position",
        )
        .bind(order_ids)
        .fetch_all(self.database.pool())
        .await
        .map_err(ApplicationError::unavailable)?;
        for row in rows {
            let order_id: Uuid = row
                .try_get("order_id")
                .map_err(ApplicationError::unavailable)?;
            items.entry(order_id).or_default().push(restore_item(&row)?);
        }
        Ok(items)
    }
}

#[async_trait]
impl OrderRepository for PostgresOrderRepository {
    async fn save(&self, order: &Order) -> Result<(), ApplicationError> {
        let mut transaction = self
            .database
            .pool()
            .begin()
            .await
            .map_err(ApplicationError::unavailable)?;
        query(
            "INSERT INTO orders (id, customer_name, status, created_at) VALUES ($1, $2, $3, $4)
             ON CONFLICT (id) DO UPDATE SET customer_name = EXCLUDED.customer_name, status = EXCLUDED.status",
        )
        .bind(order.id().as_uuid())
        .bind(order.customer_name())
        .bind(order.status().as_str())
        .bind(order.created_at())
        .execute(&mut *transaction)
        .await
        .map_err(ApplicationError::unavailable)?;
        query("DELETE FROM order_items WHERE order_id = $1")
            .bind(order.id().as_uuid())
            .execute(&mut *transaction)
            .await
            .map_err(ApplicationError::unavailable)?;
        for (position, item) in order.items().iter().enumerate() {
            query(
                "INSERT INTO order_items (id, order_id, position, product_name, quantity, unit_price, currency)
                 VALUES ($1, $2, $3, $4, $5, $6, $7)",
            )
            .bind(item.id().as_uuid())
            .bind(order.id().as_uuid())
            .bind(i32::try_from(position).map_err(ApplicationError::unavailable)?)
            .bind(item.product_name())
            .bind(item.quantity())
            .bind(item.unit_price().cents())
            .bind(item.unit_price().currency())
            .execute(&mut *transaction)
            .await
            .map_err(ApplicationError::unavailable)?;
        }
        transaction
            .commit()
            .await
            .map_err(ApplicationError::unavailable)
    }

    async fn find(&self, id: Id) -> Result<Option<Order>, ApplicationError> {
        let rows = query("SELECT id, customer_name, status, created_at FROM orders WHERE id = $1")
            .bind(id.as_uuid())
            .fetch_all(self.database.pool())
            .await
            .map_err(ApplicationError::unavailable)?;
        Ok(self.load(rows).await?.into_iter().next())
    }

    async fn list(&self, limit: usize) -> Result<Vec<Order>, ApplicationError> {
        let rows = query(
            "SELECT id, customer_name, status, created_at FROM orders ORDER BY created_at DESC LIMIT $1",
        )
            .bind(i64::try_from(limit).map_err(ApplicationError::unavailable)?)
            .fetch_all(self.database.pool())
            .await
            .map_err(ApplicationError::unavailable)?;
        self.load(rows).await
    }
}

fn restore_order(row: &PgRow, items: Vec<OrderItem>) -> Result<Order, ApplicationError> {
    let id: Uuid = row.try_get("id").map_err(ApplicationError::unavailable)?;
    let customer_name: String = row
        .try_get("customer_name")
        .map_err(ApplicationError::unavailable)?;
    let status: String = row
        .try_get("status")
        .map_err(ApplicationError::unavailable)?;
    let created_at: DateTime<Utc> = row
        .try_get("created_at")
        .map_err(ApplicationError::unavailable)?;
    let status = OrderStatus::parse(&status).map_err(ApplicationError::unavailable)?;
    Ok(Order::restore(
        Id::from_uuid(id),
        customer_name,
        status,
        created_at,
        items,
    ))
}

fn restore_item(row: &PgRow) -> Result<OrderItem, ApplicationError> {
    let id: Uuid = row.try_get("id").map_err(ApplicationError::unavailable)?;
    let product_name: String = row
        .try_get("product_name")
        .map_err(ApplicationError::unavailable)?;
    let quantity: i32 = row
        .try_get("quantity")
        .map_err(ApplicationError::unavailable)?;
    let cents: i64 = row
        .try_get("unit_price")
        .map_err(ApplicationError::unavailable)?;
    let currency: String = row
        .try_get("currency")
        .map_err(ApplicationError::unavailable)?;
    let unit_price = Money::new(cents, &currency).map_err(ApplicationError::unavailable)?;
    Ok(OrderItem::restore(
        Id::from_uuid(id),
        product_name,
        quantity,
        unit_price,
    ))
}
