use std::sync::Arc;

use axum::extract::rejection::{JsonRejection, QueryRejection};
use axum::extract::{Path, Query, State};
use axum::http::{StatusCode, header};
use axum::response::IntoResponse;
use axum::routing::get;
use axum::{Json, Router};
use chrono::SecondsFormat;
use serde::{Deserialize, Serialize};

use super::Problem;
use crate::application::sample_items::{ListSampleItems, SampleItemService, SampleItemView};
use crate::domain::shared::Id;

type Service = State<Arc<SampleItemService>>;

#[derive(Default, Deserialize)]
#[serde(default)]
struct SampleItemRequest {
    name: String,
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
struct SampleItemResponse {
    id: String,
    name: String,
    created_at: String,
}

impl From<SampleItemView> for SampleItemResponse {
    fn from(view: SampleItemView) -> Self {
        Self {
            id: view.id,
            name: view.name,
            created_at: view.created_at.to_rfc3339_opts(SecondsFormat::Secs, true),
        }
    }
}

pub fn routes(service: Arc<SampleItemService>) -> Router {
    Router::new()
        .route("/sample-items", get(list).post(create))
        .route("/sample-items/{id}", get(find).put(rename).delete(delete))
        .with_state(service)
}

async fn create(
    State(service): Service,
    payload: Result<Json<SampleItemRequest>, JsonRejection>,
) -> Result<impl IntoResponse, Problem> {
    let Json(request) = payload?;
    let id = service.create(&request.name).await?;
    Ok((
        StatusCode::CREATED,
        [(header::LOCATION, format!("/sample-items/{id}"))],
        Json(CreatedResponse { id: id.to_string() }),
    ))
}

async fn list(
    State(service): Service,
    query: Result<Query<ListQuery>, QueryRejection>,
) -> Result<Json<Vec<SampleItemResponse>>, Problem> {
    let Query(query) = query?;
    let views = service
        .list(ListSampleItems { limit: query.limit })
        .await?;
    Ok(Json(views.into_iter().map(SampleItemResponse::from).collect()))
}

async fn find(
    State(service): Service,
    Path(id): Path<String>,
) -> Result<Json<SampleItemResponse>, Problem> {
    let view = service.get(Id::parse(&id)?).await?;
    Ok(Json(SampleItemResponse::from(view)))
}

async fn rename(
    State(service): Service,
    Path(id): Path<String>,
    payload: Result<Json<SampleItemRequest>, JsonRejection>,
) -> Result<StatusCode, Problem> {
    let id = Id::parse(&id)?;
    let Json(request) = payload?;
    service.rename(id, &request.name).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn delete(State(service): Service, Path(id): Path<String>) -> Result<StatusCode, Problem> {
    service.delete(Id::parse(&id)?).await?;
    Ok(StatusCode::NO_CONTENT)
}
