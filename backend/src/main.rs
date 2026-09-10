use axum::{Router, routing::get};
use tokio::net::TcpListener;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

async fn health() -> &'static str {
    "ok"
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::from_default_env())
        .with(tracing_subscriber::fmt::layer())
        .init();

    let app = Router::new().route("/health", get(health));
    let listener = TcpListener::bind("127.0.0.1:3000").await?;

    tracing::info!(address = ?listener.local_addr()?, "BWP-SonaSea backend listening");
    axum::serve(listener, app).await?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::health;

    #[tokio::test]
    async fn health_endpoint_returns_ok() {
        assert_eq!(health().await, "ok");
    }
}
