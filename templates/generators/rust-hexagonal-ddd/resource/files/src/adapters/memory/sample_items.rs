use std::cmp::Reverse;
use std::collections::HashMap;

use async_trait::async_trait;
use tokio::sync::RwLock;

use crate::application::error::ApplicationError;
use crate::application::sample_items::SampleItemRepository;
use crate::domain::sample_items::SampleItem;
use crate::domain::shared::Id;

#[derive(Default)]
pub struct InMemorySampleItemRepository {
    sample_items: RwLock<HashMap<Id, SampleItem>>,
}

#[async_trait]
impl SampleItemRepository for InMemorySampleItemRepository {
    async fn save(&self, sample_item: &SampleItem) -> Result<(), ApplicationError> {
        self.sample_items
            .write()
            .await
            .insert(sample_item.id(), sample_item.clone());
        Ok(())
    }

    async fn find(&self, id: Id) -> Result<Option<SampleItem>, ApplicationError> {
        Ok(self.sample_items.read().await.get(&id).cloned())
    }

    async fn list(&self, limit: usize) -> Result<Vec<SampleItem>, ApplicationError> {
        let mut sample_items: Vec<SampleItem> =
            self.sample_items.read().await.values().cloned().collect();
        sample_items.sort_by_key(|sample_item| Reverse(sample_item.created_at()));
        sample_items.truncate(limit);
        Ok(sample_items)
    }

    async fn delete(&self, id: Id) -> Result<bool, ApplicationError> {
        Ok(self.sample_items.write().await.remove(&id).is_some())
    }
}
