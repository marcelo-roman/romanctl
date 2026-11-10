mod health;
mod orders;
mod problem;
// romanctl:modules

use std::sync::Arc;

use axum::Router;
use axum::routing::get;

use crate::application::orders::OrderService;
// romanctl:imports

pub use problem::Problem;

#[derive(Clone)]
pub struct Services {
    pub orders: Arc<OrderService>,
    // romanctl:services
}

pub fn router(services: Services) -> Router {
    let mut router = Router::new().route("/health", get(health::check));
    router = router.merge(orders::routes(services.orders));
    // romanctl:routes
    router
}
