use std::env;

use chrono::{SubsecRound, Utc};
use sample_app::adapters::postgres::Database;
use sample_app::adapters::postgres::orders::PostgresOrderRepository;
use sample_app::application::orders::OrderRepository;
use sample_app::domain::orders::{Money, Order, OrderStatus};
use sample_app::domain::shared::Id;

async fn database() -> Option<Database> {
    let Ok(url) = env::var("DATABASE_URL") else {
        eprintln!(
            "skipped: DATABASE_URL not set; romanctl up --no-run creates the database and the .env"
        );
        return None;
    };
    Some(
        Database::connect(&url)
            .await
            .expect("connect to DATABASE_URL"),
    )
}

#[tokio::test]
async fn order_round_trip() {
    let Some(database) = database().await else {
        return;
    };
    let repository = PostgresOrderRepository::new(database);

    let created_at = Utc::now().trunc_subsecs(6);
    let mut order = Order::create("Ana", created_at).unwrap();
    order
        .add_item("Caneta", 2, Money::new(350, "BRL").unwrap())
        .unwrap();
    order
        .add_item("Caderno", 1, Money::new(1200, "BRL").unwrap())
        .unwrap();
    repository.save(&order).await.unwrap();
    order.place(Utc::now()).unwrap();
    repository.save(&order).await.unwrap();

    let loaded = repository
        .find(order.id())
        .await
        .unwrap()
        .expect("saved order");
    assert_eq!(loaded.customer_name(), "Ana");
    assert_eq!(loaded.status(), OrderStatus::Placed);
    assert_eq!(loaded.created_at(), created_at);
    assert_eq!(loaded.items(), order.items());
    assert_eq!(loaded.total().cents(), 1900);

    let listed = repository.list(200).await.unwrap();
    assert!(listed.iter().any(|candidate| candidate.id() == order.id()));
    assert!(repository.find(Id::generate()).await.unwrap().is_none());
}

#[tokio::test(flavor = "multi_thread")]
async fn concurrent_startups_create_the_schema_once() {
    let Ok(url) = env::var("DATABASE_URL") else {
        return;
    };
    let startups = (0..4).map(|_| {
        let url = url.clone();
        tokio::spawn(async move { Database::connect(&url).await.map(|_| ()) })
    });
    for startup in startups.collect::<Vec<_>>() {
        startup.await.unwrap().expect("schema creation");
    }
}
