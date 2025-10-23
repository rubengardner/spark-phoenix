defmodule SparkPhoenix.Repo do
  use Ecto.Repo,
    otp_app: :spark_phoenix,
    adapter: Ecto.Adapters.Postgres
end
