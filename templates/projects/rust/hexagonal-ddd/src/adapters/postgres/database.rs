use sqlx::PgPool;
use sqlx::postgres::PgPoolOptions;

const SCHEMA: &str = include_str!("schema.sql");

#[derive(Clone)]
pub struct Database {
    pool: PgPool,
}

impl Database {
    pub async fn connect(url: &str) -> Result<Self, sqlx::Error> {
        let pool = PgPoolOptions::new()
            .max_connections(10)
            .connect(url)
            .await?;
        let database = Self { pool };
        database.create_schema().await?;
        Ok(database)
    }

    pub(super) fn pool(&self) -> &PgPool {
        &self.pool
    }

    async fn create_schema(&self) -> Result<(), sqlx::Error> {
        let mut transaction = self.pool.begin().await?;
        sqlx::query("SELECT pg_advisory_xact_lock(hashtext('sample_app.schema'))")
            .execute(&mut *transaction)
            .await?;
        sqlx::raw_sql(SCHEMA).execute(&mut *transaction).await?;
        transaction.commit().await
    }
}
