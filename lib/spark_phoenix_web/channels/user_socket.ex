defmodule SparkPhoenixWeb.UserSocket do
  use Phoenix.Socket

  # A true broadcast channel, allowing all clients to join the same topic.
  channel "board:lobby", SparkPhoenixWeb.BoardChannel

  @impl true
  def connect(_params, socket, _connect_info) do
    {:ok, socket}
  end

  # We don't need to identify users for this project.
  @impl true
  def id(_socket), do: nil
end
