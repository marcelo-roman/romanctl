use std::env;
use std::error::Error;

use chrono::Utc;
use sample_app::adapters::http;
use sample_app::bootstrap;
use tokio::net::TcpListener;

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    tracing_subscriber::fmt().init();
    let services = bootstrap::in_memory(Utc::now);
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
