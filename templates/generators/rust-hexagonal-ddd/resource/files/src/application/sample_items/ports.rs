use async_trait::async_trait;

use crate::application::error::ApplicationError;
use crate::domain::sample_items::SampleItem;
use crate::domain::shared::Id;

#[async_trait]
pub trait SampleItemRepository: Send + Sync {
    async fn save(&self, sample_item: &SampleItem) -> Result<(), ApplicationError>;
    async fn find(&self, id: Id) -> Result<Option<SampleItem>, ApplicationError>;
    async fn list(&self, limit: usize) -> Result<Vec<SampleItem>, ApplicationError>;
    async fn delete(&self, id: Id) -> Result<bool, ApplicationError>;
}
