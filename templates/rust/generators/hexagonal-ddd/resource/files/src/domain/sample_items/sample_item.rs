use chrono::{DateTime, Utc};

use crate::domain::shared::{Id, RuleViolation};

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct SampleItem {
    id: Id,
    name: String,
    created_at: DateTime<Utc>,
}

impl SampleItem {
    pub fn create(name: &str, now: DateTime<Utc>) -> Result<Self, RuleViolation> {
        Ok(Self {
            id: Id::generate(),
            name: required_name(name)?,
            created_at: now,
        })
    }

    pub fn restore(id: Id, name: String, created_at: DateTime<Utc>) -> Self {
        Self {
            id,
            name,
            created_at,
        }
    }

    pub fn id(&self) -> Id {
        self.id
    }

    pub fn name(&self) -> &str {
        &self.name
    }

    pub fn created_at(&self) -> DateTime<Utc> {
        self.created_at
    }

    pub fn rename(&mut self, name: &str) -> Result<(), RuleViolation> {
        self.name = required_name(name)?;
        Ok(())
    }
}

fn required_name(name: &str) -> Result<String, RuleViolation> {
    let name = name.trim();
    if name.is_empty() {
        return Err(RuleViolation::new("O nome é obrigatório."));
    }
    Ok(name.to_owned())
}
