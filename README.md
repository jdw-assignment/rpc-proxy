## Write Up

### Application

#### Endpoints

| HTTP Method | URL     | Public |
|-------------|---------|--------|
| POST        | /rpc    | True   |
| GET         | /health | False  |

- `/rpc` is exposed publically and proxies the rpc requests to the upstream providers.
- `/health` is an internal endpoint used by ECS to determine the status of the containers.

#### RPC Proxy

[<img width=600 src=".github/imgs/request_flow.png?raw=true">](.github/imgs/request_flow.png?raw=true)

As shown in the diagram above, incoming RPC requests are decoded by the proxy app. This decoding is required to allow the app to verify whether the requested RPC method is in the allowed list. If the method is blocked, an error response is sent, indicating that the RPC method is not allowed.

For allowed methods, the request is proxied to the RPC provider. To ensure minimal latency, responses from the provider are returned without any decoding.

However, this approach assumes the provider’s response is trustworthy, which might not always be feasible. Depending on the use case, decoding responses to verify data validity could be necessary.

#### Unit Testing

- `Test_ProxyRPCRequest_BlockedMethods`: This test checks that all RPC methods, other than `eth_blockNumber` and `eth_getBlockByNumber`, return an error.
- `Test_ProxyRPCRequest_AllowedMethods`: This test checks that both `eth_blockNumber` and `eth_getBlockByNumber` are allowed.
- `Test_ProxyRPCRequest_Proxy`: This test checks that proxying with a mock endpoint returns a valid response.

#

### Terraform / AWS Architecture

[<img width=500 src=".github/imgs/aws.png?raw=true">](.github/imgs/aws.png?raw=true)

For this homework exercise, `ap-southeast-1` and it's three available
zone (`ap-southeast-1a`, `ap-southeast-1b`, `ap-southeast-1c`) are used.

As per the requirements, ECS with fargate is being used to deploy the proxy application. An ALB is implemented to route
traffic between the different AZs.

Logs from the applications are also piped to CloudWatch. It may be a bit overzealous for this application, but logs of
requests handled by the ALB are stored in an S3 bucket.

### Terraform

As for IaC, Terraform is used as per requirement. For this assignment, Terraform AWS modules by
`terraform-aws-modules` are utilized mainly because they abstract the creation and configuration of resources.

Autoscaling has also implemented to auto scale based on the CPU load.

Custom modules were created for AWS region. In the event that multi-region deployment is needed, these modules can be
used to quickly replicate the infrastructure in a new region.

```terraform
module "ap_southeast_1" {
  source = "./aws/region"

  providers = {
    aws = aws.ap-southeast-1
  }
}

// New Region
module "us-east-1" {
  source = "./aws/region"

  providers = {
    aws = aws.us-east-1
  }
}

```

#

### CI/CD

A GitHub workflow for unit testing is implemented to verify the success of the proxy application’s build and unit tests
when a pull request is created and when commits are made to the main branch.

Similarly, when a release is made on GitHub, another workflow is executed to build the application container image and upload
it to GHCR.

#

### Observability (OpenTelemetry)

For this exercise and due to time constraints, a basic OpenTelemetry implementation is added to the web application to
log and trace incoming requests, as well as to capture any errors encountered while proxying the RPC requests. The
output is captured in CloudWatch, providing centralized and accessible monitoring and logging capabilities.

---

## Improvement for production

The following are the changes which I think will be needed for a production system.

### Multi Region

[<img width=500 src=".github/imgs/multiregion.png?raw=true">](.github/imgs/multiregion.png?raw=true)

Assuming that this service is designed to cater to users globally, this architecture could be replicated in different
regions. However, some changes would be needed to accommodate global usage. Specifically, the implementation of AWS
Global Accelerator could be used to optimally route traffic to the appropriate regional ALB endpoint which is closest to the
origin of the requests.

Furthermore, deploying the service in multiple regions increases not only the performance of the proxy service but also
its availability. In the event that one region goes down or experiences a very high load, traffic can be seamlessly
routed to other regions

#

### CI/CD

Terraform Cloud can be implemented for centralized state management and deployments to streamline and automate
infrastructure
management.

A GitHub workflow can be configured to integrate with Terraform Cloud, automatically triggering it to generate a plan
for any proposed changes upon a release made.

#

### Caching

A caching service could be implemented to cache the RPC results. This would decrease the latency of requests as it does
not need to be proxied to the RPC providers, reducing one external network call.

#

### Cloudflare

Cloudflare could also be implemented for its WAF and CDN capabilities. While AWS offers similar services with AWS WAF
and Shield, my experience suggests that AWS Shield/WAF is quite limited in its customizability. For instance, Cloudflare
allows for rate limiting with periods as short as 10 seconds, whereas AWS WAF has a minimum period of 60 seconds, which
may not be feasible for production environments.

#### WAF

As this is a public API, bots may misuse it. Implementing Cloudflare’s Super Bot feature could help reduce the number of
bot requests. However, this may not be feasible if the API service is intended for use by bots. In that scenario, rate
limiting could be utilized instead.

#### Rate-limiting

Rate limiting rules could be established to ensure that the API server is not overwhelmed by excessive requests.

#

### Multiple RPC Providers (Availability)

Currently, only one RPC provider is utilized, which may cause availability issues if the provider goes down. To mitigate
this, another service should be implemented to serve as a gateway to multiple RPC providers.

Features of the rpc gateway service should include:

1. **Provider List Management**: Maintain a list of multiple RPC providers, ensuring that there are always
   alternative providers available in case one goes down.
2. **Health Checks**: Regularly probe each RPC provider’s endpoint to check for latency and availability. This helps in
   identifying the best provider to route requests to, based on current performance and uptime.
3. **Failover Mechanism**: Automatically switch to an alternative provider if the current one becomes unavailable or
   experiences high latency. This ensures continuous service availability and minimizes downtime.
4. **Monitoring and Alerts**: Implement monitoring and alerting mechanisms to notify administrators of any issues with
   the RPC providers, allowing for quick response and resolution.

#

### Observability

Proper observability must be implemented. Using OTel, requests and errors encountered by the services. This enables
comprehensive monitoring and analysis of the system’s behavior. By logging detailed information about requests and
errors, the team can quickly identify and diagnose issues and improved system reliability.

#

### Metrics & Alarms

CloudWatch alarms or an external monitoring service should be implemented to notify the team if any services are down or
experiencing high latency, etc. This allows the team to manually intervene when necessary.

With appropriate metrics logged, you can gain insights into the behavior and performance of the applications, allowing
for further optimization of the service. For example, if metrics determine that `eth_blockNumber` has very high usage, a
caching service could be implemented to always cache the latest result.

#

### Domain and HTTPS

Before moving to production, it is essential to set up a domain and obtain HTTPS certification. Using the ALB endpoint directly is impractical. An HTTPS certificate ensures that all data transmitted between users and the server is encrypted, which is critical for maintaining the security of sensitive information.

Furthermore, using a recognizable domain associated with a known brand, combined with an HTTPS certificate, significantly enhances user trust in the service.

#

### Autoscaling

More auto-scaling options can be set up based on different criteria, for example, the number of requests hitting the ALB. This ensures that the auto-scaling strategies cater to different conditions and provide optimal performance and cost management.​⬤

#
