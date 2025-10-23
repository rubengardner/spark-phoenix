# THis is cool. This is a very cool typing system.
#

defmodule SparkPhoenixWeb.SparkController do
  use SparkPhoenixWeb, :controller
  alias SparkPhoenix.SparkEvent

  @spec create(Plug.Conn.t(), %{"x" => String.t(), "y" => String.t()}) :: Plug.Conn.t()
  def create(conn, %{"x" => x, "y" => y}) do
    params =
      conn.body_params
      |> Map.merge(%{"x" => x, "y" => y})
      |> Map.new(fn {k, v} -> {String.to_atom(k), v} end)

    case SparkEvent.changeset(params) do
      %Ecto.Changeset{valid?: true} = changeset ->
        spark_event = Ecto.Changeset.apply_changes(changeset)

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
