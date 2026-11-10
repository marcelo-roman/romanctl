use std::sync::Arc;

use axum::extract::rejection::{JsonRejection, QueryRejection};
use axum::extract::{Path, Query, State};
use axum::http::{StatusCode, header};
use axum::response::IntoResponse;
use axum::routing::{get, post};
use axum::{Json, Router};
use chrono::SecondsFormat;
use serde::{Deserialize, Serialize};

use super::Problem;
use crate::application::orders::{AddItem, ListOrders, OrderItemView, OrderService, OrderView};
use crate::domain::shared::Id;

type Service = State<Arc<OrderService>>;

#[derive(Default, Deserialize)]
#[serde(default, rename_all = "camelCase")]
struct CreateOrderRequest {
    customer_name: String,
}

#[derive(Default, Deserialize)]
#[serde(default, rename_all = "camelCase")]
struct AddItemRequest {
    product_name: String,
    quantity: i32,
    unit_price: f64,
    currency: Option<String>,
}

#[derive(Deserialize)]
struct ListQuery {
    limit: Option<usize>,
}

#[derive(Serialize)]
struct CreatedResponse {
    id: String,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct OrderItemResponse {
    product_name: String,
    quantity: i32,
    unit_price: f64,
    subtotal: f64,
}

#[derive(Serialize)]
#[serde(rename_all = "camelCase")]
struct OrderResponse {
    id: String,
    customer_name: String,
    status: String,
    created_at: String,
    total: f64,
    currency: String,
    items: Vec<OrderItemResponse>,
}

impl From<OrderItemView> for OrderItemResponse {
    fn from(view: OrderItemView) -> Self {
        Self {
            product_name: view.product_name,
            quantity: view.quantity,
            unit_price: view.unit_price,
            subtotal: view.subtotal,
        }
    }
}

impl From<OrderView> for OrderResponse {
    fn from(view: OrderView) -> Self {
        Self {
            id: view.id,
            customer_name: view.customer_name,
            status: view.status,
            created_at: view.created_at.to_rfc3339_opts(SecondsFormat::Secs, true),
            total: view.total,
            currency: view.currency,
            items: view
                .items
                .into_iter()
                .map(OrderItemResponse::from)
                .collect(),
        }
    }
}

pub fn routes(service: Arc<OrderService>) -> Router {
    Router::new()
        .route("/orders", post(create).get(list))
        .route("/orders/{id}", get(find))
        .route("/orders/{id}/items", post(add_item))
        .route("/orders/{id}/place", post(place))
        .with_state(service)
}

async fn create(
    State(service): Service,
    payload: Result<Json<CreateOrderRequest>, JsonRejection>,
) -> Result<impl IntoResponse, Problem> {
    let Json(request) = payload?;
    let id = service.create(&request.customer_name).await?;
    Ok((
        StatusCode::CREATED,
        [(header::LOCATION, format!("/orders/{id}"))],
        Json(CreatedResponse { id: id.to_string() }),
    ))
}

async fn add_item(
    State(service): Service,
    Path(id): Path<String>,
    payload: Result<Json<AddItemRequest>, JsonRejection>,
) -> Result<StatusCode, Problem> {
    let id = Id::parse(&id)?;
    let Json(request) = payload?;
    let command = AddItem {
        product_name: request.product_name,
        quantity: request.quantity,
        unit_price: request.unit_price,
        currency: request.currency,
    };
    service.add_item(id, command).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn place(State(service): Service, Path(id): Path<String>) -> Result<StatusCode, Problem> {
    service.place(Id::parse(&id)?).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn find(
    State(service): Service,
    Path(id): Path<String>,
) -> Result<Json<OrderResponse>, Problem> {
    let view = service.get(Id::parse(&id)?).await?;
    Ok(Json(OrderResponse::from(view)))
}

async fn list(
    State(service): Service,
    query: Result<Query<ListQuery>, QueryRejection>,
) -> Result<Json<Vec<OrderResponse>>, Problem> {
    let Query(query) = query?;
    let views = service.list(ListOrders { limit: query.limit }).await?;
    Ok(Json(views.into_iter().map(OrderResponse::from).collect()))
}
