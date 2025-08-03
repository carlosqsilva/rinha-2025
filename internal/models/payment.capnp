@0xcfa902c41d794181;

using Go = import "/go.capnp";

$Go.package("models");
$Go.import("internal/models");

struct PaymentData {
  correlationId @0 :Text;
  amount        @1 :Float64;
  requestedAt   @2 :Text;
  processor     @3 :UInt8 = 0;
}
