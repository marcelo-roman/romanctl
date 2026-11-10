use chrono::{DateTime, Utc};

pub type Clock = fn() -> DateTime<Utc>;
