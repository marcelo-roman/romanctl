mod support;

use axum::http::StatusCode;
use support::{app, send};

#[tokio::test]
async fn sample_item_crud() {
    let app = app();
    let (status, created) = send(&app, "POST", "/sample-items", r#"{"name":"Primeiro"}"#).await;
    assert_eq!(status, StatusCode::CREATED);
    let uri = format!("/sample-items/{}", created["id"].as_str().expect("id"));

    let (status, _) = send(&app, "PUT", &uri, r#"{"name":"Segundo"}"#).await;
    assert_eq!(status, StatusCode::NO_CONTENT);
    let (status, found) = send(&app, "GET", &uri, "").await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(found["name"], "Segundo");
    assert!(found["createdAt"].is_string());

    let (status, list) = send(&app, "GET", "/sample-items?limit=10", "").await;
    assert_eq!(status, StatusCode::OK);
    assert_eq!(list.as_array().map(Vec::len), Some(1));

    let (status, _) = send(&app, "DELETE", &uri, "").await;
    assert_eq!(status, StatusCode::NO_CONTENT);
    let (status, _) = send(&app, "GET", &uri, "").await;
    assert_eq!(status, StatusCode::NOT_FOUND);
    let (status, _) = send(&app, "DELETE", &uri, "").await;
    assert_eq!(status, StatusCode::NOT_FOUND);
}

#[tokio::test]
async fn sample_item_errors_become_problems() {
    let app = app();
    let (status, problem) = send(&app, "POST", "/sample-items", r#"{"name":" "}"#).await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    assert_eq!(problem["detail"], "O nome é obrigatório.");
    let (status, _) = send(&app, "POST", "/sample-items", "{").await;
    assert_eq!(status, StatusCode::BAD_REQUEST);
    let (status, problem) = send(&app, "GET", "/sample-items/abc", "").await;
    assert_eq!(status, StatusCode::UNPROCESSABLE_ENTITY);
    assert_eq!(problem["detail"], "Identificador inválido.");
}
