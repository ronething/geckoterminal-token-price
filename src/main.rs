use std::collections::HashMap;
use std::env;
use std::fs;
use std::time::SystemTime;
use chrono::{DateTime, Utc};
use reqwest::blocking::Client;
use serde::Deserialize;
use serde_json::json;

#[derive(Deserialize)]
struct GetTokenPriceResp {
    data: TokenPriceData,
}

#[derive(Deserialize)]
struct TokenPriceData {
    attributes: TokenPriceAttributes,
}

#[derive(Deserialize)]
struct TokenPriceAttributes {
    token_prices: HashMap<String, String>,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let ding_token = env::var("DINGTALK_TOKEN").expect("DINGTALK_TOKEN not set");

    let token_list = fs::read_to_string("token")?;
    let tokens: Vec<&str> = token_list.trim().split('\n').collect();

    let mut network_addrs: HashMap<String, Vec<String>> = HashMap::new();
    let mut token_name: HashMap<String, String> = HashMap::new();

    for token in tokens {
        let parts: Vec<&str> = token.split(',').collect();
        network_addrs.entry(parts[0].to_string())
            .or_insert_with(Vec::new)
            .push(parts[1].to_string());
        token_name.insert(parts[1].to_string(), parts[2].to_string());
    }

    let client = Client::new();
    let mut addr_price: HashMap<String, String> = HashMap::new();

    for (network, addrs) in network_addrs {
        let url = format!("https://api.geckoterminal.com/api/v2/simple/networks/{}/token_price/{}", 
                          network, addrs.join(","));
        let resp: GetTokenPriceResp = client.get(&url)
            .header("Accept", "application/json;version=20230302")
            .send()?
            .json()?;

        addr_price.extend(resp.data.attributes.token_prices);
    }

    let now: DateTime<Utc> = SystemTime::now().into();
    let mut send_text = format!("token price, time: {}\n", now.to_rfc3339());

    for (addr, price) in addr_price {
        send_text.push_str(&format!("name: {}, addr: {}, price: {}\n", 
                                    token_name.get(&addr).unwrap_or(&"Unknown".to_string()), 
                                    addr, price));
    }

    // Send message to DingTalk
    let ding_url = format!("https://oapi.dingtalk.com/robot/send?access_token={}", ding_token);
    let payload = json!({
        "msgtype": "text",
        "text": {
            "content": send_text
        }
    });

    let response = client.post(&ding_url)
        .header("Content-Type", "application/json")
        .json(&payload)
        .send()?;

    if response.status().is_success() {
        println!("Message sent successfully");
    } else {
        println!("Message failed to send: {:?}", response.text()?);
    }

    Ok(())
}
