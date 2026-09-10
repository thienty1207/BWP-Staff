use backend::{admin::seed::seed_development_admin, app, config::Config, shared::database};
use tokio::net::TcpListener;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenvy::dotenv().ok();

    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::from_default_env())
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = Config::from_env()?;
    let pool = database::connect(config.database()).await?;
    database::run_migrations(&pool).await?;

    if let Some(admin) = config.seed_admin() {
        let outcome = seed_development_admin(&pool, admin).await?;
        tracing::info!(?outcome, "development seed checked");
    }

    let app = app::router();
    let listener = TcpListener::bind("127.0.0.1:3000").await?;

    tracing::info!(address = ?listener.local_addr()?, "BWP-SonaSea backend listening");
    axum::serve(listener, app).await?;

    Ok(())
}
