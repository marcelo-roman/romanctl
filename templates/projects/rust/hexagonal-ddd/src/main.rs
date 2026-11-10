use std::env;
use std::error::Error;

use chrono::Utc;
use sample_app::adapters::http;
use sample_app::bootstrap;
use tokio::net::TcpListener;

const DEFAULT_DATABASE_URL: &str = "postgres://postgres:postgres@127.0.0.1:5432/sample_app";

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    tracing_subscriber::fmt().init();
    let url = env::var("DATABASE_URL").unwrap_or_else(|_| DEFAULT_DATABASE_URL.to_owned());
    let services = bootstrap::postgres(&url, Utc::now).await?;
    let port = env::var("PORT").unwrap_or_else(|_| "3100".to_owned());
    let listener = TcpListener::bind(format!("0.0.0.0:{port}")).await?;
    tracing::info!("listening on http://localhost:{port}");
    axum::serve(listener, http::router(services))
        .with_graceful_shutdown(shutdown())
        .await?;
    Ok(())
}

async fn shutdown() {
    let interrupt = tokio::signal::ctrl_c();
    #[cfg(unix)]
    let terminate = async {
        if let Ok(mut signal) =
            tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
        {
            signal.recv().await;
        }
    };
    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();
    tokio::select! {
        _ = interrupt => {}
        _ = terminate => {}
    }
}
