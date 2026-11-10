mod support;

use axum::http::StatusCode;
use support::{app, send};

#[tokio::test]
async fn order_lifecycle() {
    let app = app();
    let (status, created) = send(&app, "POST", "/orders", r#"{"customerName":"Ana"}"#).await;
    assert_eq!(status, StatusCode::CREATED);
    let id = created["id"].as_str().expect("id").to_owned();

    let item = r#"{"productName":"Caneta","quantity":2,"unitPrice":3.5}"#;
    let (status, _) = send(&app, "POST", &format!("/orders/{id}/items"), item).await;
    assert_eq!(status, StatusCode::NO_CONTENT);
    let (status, _) = send(&app, "POST", &format!("/orders/{id}/place"), "").await;
    assert_eq!(status, StatusCode::NO_CONTENT);

    let (status, order) = send(&app, "GET", &format!("/orders/{id}"), "").await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(order["status"], "Placed");
    assert_eq!(order["customerName"], "Ana");
    assert_eq!(order["total"], 7.0);
    assert_eq!(order["currency"], "BRL");
    assert_eq!(order["items"][0]["productName"], "Caneta");
    assert_eq!(order["items"][0]["unitPrice"], 3.5);
    assert!(order["createdAt"].is_string());

    let (status, list) = send(&app, "GET", "/orders?limit=10", "").await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(list.as_array().map(Vec::len), Some(1));
}

#[tokio::test]
async fn errors_become_problems() {
    let app = app();
    let (status, problem) = send(&app, "POST", "/orders", r#"{"customerName":""}"#).await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    assert_eq!(problem["status"], 422);
    assert_eq!(problem["detail"], "O nome do cliente é obrigatório.");

    let missing = "/orders/00000000-0000-4000-8000-000000000000";
    let (status, problem) = send(&app, "GET", missing, "").await;
    assert_eq!(status, StatusCode::NOT_FOUND);
    assert_eq!(problem["title"], "Não encontrado");

    let (status, _) = send(&app, "POST", "/orders", "{").await;
    assert_eq!(status, StatusCode::BAD_REQUEST);

    let (status, problem) = send(&app, "GET", "/orders/abc", "").await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    assert_eq!(problem["detail"], "Identificador inválido.");
}

#[tokio::test]
async fn placing_an_empty_order_is_rejected() {
    let app = app();
    let (_, created) = send(&app, "POST", "/orders", r#"{"customerName":"Ana"}"#).await;
    let id = created["id"].as_str().expect("id").to_owned();
    let (status, problem) = send(&app, "POST", &format!("/orders/{id}/place"), "").await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    assert_eq!(
        problem["detail"],
        "Um pedido sem itens não pode ser realizado."
    );
}

#[tokio::test]
async fn health_answers_ok() {
    let (status, body) = send(&app(), "GET", "/health", "").await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(body["status"], "ok");
}
