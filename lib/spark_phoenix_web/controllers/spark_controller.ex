# THis is cool. This is a very cool typing system.
#

defmodule SparkPhoenixWeb.SparkController do
  use SparkPhoenixWeb, :controller
  alias SparkPhoenix.SparkEvent

  @doc """
  Receives a spark event, validates parameters, and broadcasts it to clients.
  """
  @spec create(Plug.Conn.t(), map()) :: Plug.Conn.t()
  def create(conn, %{"x" => x, "y" => y}) do
    # Combine path params and body params, and convert keys to atoms for the changeset
    params =
      conn.body_params
      |> Map.merge(%{"x" => x, "y" => y})
      |> Map.new(fn {k, v} -> {String.to_atom(k), v} end)

    case SparkEvent.changeset(params) do
      %Ecto.Changeset{valid?: true} = changeset ->
        spark_event = Ecto.Changeset.apply_changes(changeset)

        # US-03: Broadcast the event asynchronously so the client gets a fast response.
        Task.start(fn ->
          SparkPhoenixWeb.Endpoint.broadcast("board:lobby", "sparkle_event", %{data: spark_event})
        end)

        conn
        |> put_status(:accepted)
        |> json(%{status: "accepted", data: spark_event})

      %Ecto.Changeset{valid?: false} = changeset ->
        conn
        |> put_status(:bad_request)
        |> json(%{
          error: "Invalid parameters",
          details: Ecto.Changeset.traverse_errors(changeset, &elem(&1, 0))
        })
    end
  end
end
