
#[macro_use]
extern crate lazy_static;

use actix_cors::Cors;
use actix_web::{get, post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
struct JWK{
    kty: String,
    k: String,
    kid: String,
    alg: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct JWKArray{
    keys: Vec<JWK>,
}


#[derive(Debug, Deserialize, Serialize)]
struct QueryPrint{
    text: String,
}

#[get("/")]
async fn home() -> impl Responder {
    HttpResponse::Ok()
}

#[get("/ping")]
async fn ping() -> impl Responder {
    println!("hello world!");
    HttpResponse::Ok().body("Hello world!")
}

#[post("/print")]
async fn print(msg: web::Query<QueryPrint>) -> impl Responder {
    let msg = msg.0.text;
    println!("message: {}", msg.clone());
    HttpResponse::Ok().body(msg)
}

#[get("/jwk")]
async fn jwk() -> impl Responder {
    let keys = JWKArray{
        keys: vec![
            JWK{
                kty: "oct".to_string(),
                k: "AyM1SysPpbyDfgZld3umj1qzKObwVMkoqQ-EstJQLr_T-1qS0gZH75aKtMN3Yj0iPS4hcgUuTwjAzZr1Z9CAow".to_string(),
                kid: "sim2".to_string(),
                alg: "HS256".to_string(),
            },
        ],
    };
    HttpResponse::Ok().json(keys)
}


#[actix_web::main]
async fn main() -> std::io::Result<()> {
    dotenv::dotenv().ok();
    HttpServer::new(move|| {
        let cors = Cors::permissive();
        App::new()
            .wrap(cors)
            .service(home)
            .service(ping)
            .service(print)
            .service(jwk)
    })
    .bind("0.0.0.0:8000")?
    .run()
    .await
}