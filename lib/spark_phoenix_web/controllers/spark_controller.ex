defmodule SparkPhoenixWeb.SparkController do
  use SparkPhoenixWeb, :controller

  @doc """
  Receives a spark event, validates parameters, and prepares for broadcast.
  """
  @spec create(Plug.Conn.t(), %{"x" => String.t(), "y" => String.t()}) :: Plug.Conn.t()
  def create(conn, %{"x" => x, "y" => y}) do
    # US-02: Parse and validate parameters from the JSON body.
    with %{
           "color" => color,
           "radius" => radius,
           "transparency" => transparency,
           "time_to_grow" => time_to_grow
         } <- conn.body_params,
         is_binary(color),
         is_integer(radius) and radius > 0,
         is_number(transparency) and transparency >= 0.0 and transparency <= 1.0,
         is_integer(time_to_grow) and time_to_grow > 0 do
      # Parameters are valid. For now, just return them in the response.
      # Broadcasting will be handled in a later step.
      conn
      |> put_status(:accepted)
      |> json(%{
        status: "accepted",
        x: x,
        y: y,
        color: color,
        radius: radius,
        transparency: transparency,
        time_to_grow: time_to_grow
      })
    else
      # Handle cases where parameters are missing, have the wrong type, or the body is invalid.
      _error ->
        conn
        |> put_status(:bad_request)
        |> json(%{
          error: "Invalid or missing spark parameters in JSON body.",
          reason:
            "Check types and ranges (color: string, radius: int > 0, transparency: num 0-1, time_to_grow: int > 0)."
        })
    end
  end
end
