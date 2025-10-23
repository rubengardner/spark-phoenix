defmodule SparkPhoenixWeb.BoardChannel do
  use SparkPhoenixWeb, :channel

  @impl true
  def join("board:lobby", _payload, socket) do
    # All clients join the same "board:lobby" topic.
    {:ok, socket}
  end

  # We don't need to handle incoming messages from the client for this channel.
  # The server only broadcasts to it.
  @impl true
  def handle_in(_event, _payload, socket) do
    {:noreply, socket}
  end
end
