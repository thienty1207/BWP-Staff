use sqlx::{PgPool, postgres::PgPoolOptions};

use crate::config::DatabaseSettings;

pub async fn connect(settings: &DatabaseSettings) -> Result<PgPool, sqlx::Error> {
    PgPoolOptions::new()
        .min_connections(settings.min_connections())
        .max_connections(settings.max_connections())
        .acquire_timeout(settings.acquire_timeout())
        .connect(settings.url())
        .await
}

pub async fn run_migrations(pool: &PgPool) -> Result<(), sqlx::migrate::MigrateError> {
    sqlx::migrate!("./migrations").run(pool).await
}
