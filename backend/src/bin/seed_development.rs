use std::io;

use backend::{config::Config, seed::seed_development_admin, shared::database};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenvy::dotenv().ok();

    let config = Config::from_env()?;
    let admin = config.seed_admin().ok_or_else(|| {
        io::Error::new(
            io::ErrorKind::InvalidInput,
            "development seed is disabled; set APP_ENV=development and SEED_DEVELOPMENT_DATA=true",
        )
    })?;

    let pool = database::connect(config.database()).await?;
    database::run_migrations(&pool).await?;
    let outcome = seed_development_admin(&pool, admin).await?;

    println!("development database initialized: {outcome:?}");
    pool.close().await;

    Ok(())
}
