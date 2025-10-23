defmodule SparkPhoenixWeb.HealthController do
  use SparkPhoenixWeb, :controller

  def index(conn, _params) do
    json(conn, %{status: "ok"})
  end
end
