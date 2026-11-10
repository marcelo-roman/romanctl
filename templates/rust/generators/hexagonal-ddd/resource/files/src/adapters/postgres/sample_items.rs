use async_trait::async_trait;
use chrono::{DateTime, Utc};
use sqlx::postgres::PgRow;
use sqlx::{Row, query};
use uuid::Uuid;

use super::Database;
use crate::application::error::ApplicationError;
use crate::application::sample_items::SampleItemRepository;
use crate::domain::sample_items::SampleItem;
use crate::domain::shared::Id;

pub struct PostgresSampleItemRepository {
    database: Database,
}

impl PostgresSampleItemRepository {
    pub fn new(database: Database) -> Self {
        Self { database }
    }
}

#[async_trait]
impl SampleItemRepository for PostgresSampleItemRepository {
    async fn save(&self, sample_item: &SampleItem) -> Result<(), ApplicationError> {
        query(
            "INSERT INTO sample_items (id, name, created_at) VALUES ($1, $2, $3)
             ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name",
        )
        .bind(sample_item.id().as_uuid())
        .bind(sample_item.name())
        .bind(sample_item.created_at())
        .execute(self.database.pool())
        .await
        .map_err(ApplicationError::unavailable)?;
        Ok(())
    }

    async fn find(&self, id: Id) -> Result<Option<SampleItem>, ApplicationError> {
        let row = query("SELECT id, name, created_at FROM sample_items WHERE id = $1")
            .bind(id.as_uuid())
            .fetch_optional(self.database.pool())
            .await
            .map_err(ApplicationError::unavailable)?;
        row.as_ref().map(restore).transpose()
    }

    async fn list(&self, limit: usize) -> Result<Vec<SampleItem>, ApplicationError> {
        let rows = query(
            "SELECT id, name, created_at FROM sample_items ORDER BY created_at DESC LIMIT $1",
        )
        .bind(i64::try_from(limit).map_err(ApplicationError::unavailable)?)
        .fetch_all(self.database.pool())
        .await
        .map_err(ApplicationError::unavailable)?;
        rows.iter().map(restore).collect()
    }

    async fn delete(&self, id: Id) -> Result<bool, ApplicationError> {
        let result = query("DELETE FROM sample_items WHERE id = $1")
            .bind(id.as_uuid())
            .execute(self.database.pool())
            .await
            .map_err(ApplicationError::unavailable)?;
        Ok(result.rows_affected() > 0)
    }
}

fn restore(row: &PgRow) -> Result<SampleItem, ApplicationError> {
    let id: Uuid = row.try_get("id").map_err(ApplicationError::unavailable)?;
    let name: String = row.try_get("name").map_err(ApplicationError::unavailable)?;
    let created_at: DateTime<Utc> = row
        .try_get("created_at")
        .map_err(ApplicationError::unavailable)?;
    Ok(SampleItem::restore(Id::from_uuid(id), name, created_at))
}
