use chrono::{DateTime, Utc};

use crate::domain::sample_items::SampleItem;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SampleItemView {
    pub id: String,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

impl From<&SampleItem> for SampleItemView {
    fn from(sample_item: &SampleItem) -> Self {
        Self {
            id: sample_item.id().to_string(),
            name: sample_item.name().to_owned(),
            created_at: sample_item.created_at(),
        }
    }
}
