defmodule SparkPhoenixWeb.PageController do
  use SparkPhoenixWeb, :controller

  def home(conn, _params) do
    render(conn, :home)
  end
end
