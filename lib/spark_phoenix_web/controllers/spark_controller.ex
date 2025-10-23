defmodule SparkPhoenixWeb.SparkController do
  use SparkPhoenixWeb, :controller

  def create(conn, %{"x" => x, "y" => y}) do
    # We'll add body parsing and broadcasting later.
    # For now, just confirm the endpoint works.
    conn
    |> put_status(:accepted)
    |> json(%{status: "accepted", x: x, y: y})
  end
end
