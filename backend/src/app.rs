use axum::{Router, routing::get};

/// Build the application router.
///
/// Feature routers are mounted here as their specifications become
/// implementable. Keeping this composition point separate from `main` makes
/// the HTTP surface easy to test and keeps startup concerns isolated.
pub fn router() -> Router {
    Router::new().route("/health", get(health))
}

async fn health() -> &'static str {
    "ok"
}

#[cfg(test)]
mod tests {
    use super::health;

    #[tokio::test]
    async fn health_endpoint_returns_ok() {
        assert_eq!(health().await, "ok");
    }
}
