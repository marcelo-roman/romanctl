use std::sync::Arc;

use super::{ListSampleItems, SampleItemRepository, SampleItemView};
use crate::application::clock::Clock;
use crate::application::error::ApplicationError;
use crate::domain::sample_items::SampleItem;
use crate::domain::shared::Id;

const NOT_FOUND: &str = "SampleItem não encontrado.";

pub struct SampleItemService {
    repository: Arc<dyn SampleItemRepository>,
    clock: Clock,
}

impl SampleItemService {
    pub fn new(repository: Arc<dyn SampleItemRepository>, clock: Clock) -> Self {
        Self { repository, clock }
    }

    pub async fn create(&self, name: &str) -> Result<Id, ApplicationError> {
        let sample_item = SampleItem::create(name, (self.clock)())?;
        self.repository.save(&sample_item).await?;
        Ok(sample_item.id())
    }

    pub async fn rename(&self, id: Id, name: &str) -> Result<(), ApplicationError> {
        let mut sample_item = self.load(id).await?;
        sample_item.rename(name)?;
        self.repository.save(&sample_item).await
    }

    pub async fn get(&self, id: Id) -> Result<SampleItemView, ApplicationError> {
        let sample_item = self.load(id).await?;
        Ok(SampleItemView::from(&sample_item))
    }

    pub async fn list(
        &self,
        query: ListSampleItems,
    ) -> Result<Vec<SampleItemView>, ApplicationError> {
        let sample_items = self.repository.list(query.limit()).await?;
        Ok(sample_items.iter().map(SampleItemView::from).collect())
    }

    pub async fn delete(&self, id: Id) -> Result<(), ApplicationError> {
        if !self.repository.delete(id).await? {
            return Err(ApplicationError::not_found(NOT_FOUND));
        }
        Ok(())
    }

    async fn load(&self, id: Id) -> Result<SampleItem, ApplicationError> {
        self.repository
            .find(id)
            .await?
            .ok_or_else(|| ApplicationError::not_found(NOT_FOUND))
    }
}
