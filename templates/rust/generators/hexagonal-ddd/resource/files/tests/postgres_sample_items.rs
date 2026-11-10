use std::env;

use chrono::{SubsecRound, Utc};
use sample_app::adapters::postgres::Database;
use sample_app::adapters::postgres::sample_items::PostgresSampleItemRepository;
use sample_app::application::sample_items::SampleItemRepository;
use sample_app::domain::sample_items::SampleItem;

#[tokio::test]
async fn sample_item_round_trip() {
    let Ok(url) = env::var("DATABASE_URL") else {
        eprintln!("skipped: DATABASE_URL not set; romanctl up sample-app --no-run creates the database and the .env");
        return;
    };
    let database = Database::connect(&url).await.expect("connect to DATABASE_URL");
    let repository = PostgresSampleItemRepository::new(database);

    let mut sample_item = SampleItem::create("Primeiro", Utc::now().trunc_subsecs(6)).unwrap();
    repository.save(&sample_item).await.unwrap();
    sample_item.rename("Segundo").unwrap();
    repository.save(&sample_item).await.unwrap();

    let loaded = repository.find(sample_item.id()).await.unwrap();
    assert_eq!(loaded.as_ref(), Some(&sample_item));
    let listed = repository.list(200).await.unwrap();
    assert!(listed.contains(&sample_item));

    assert!(repository.delete(sample_item.id()).await.unwrap());
    assert!(!repository.delete(sample_item.id()).await.unwrap());
    assert!(repository.find(sample_item.id()).await.unwrap().is_none());
}
