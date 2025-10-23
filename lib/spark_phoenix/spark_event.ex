defmodule SparkPhoenix.SparkEvent do
  @moduledoc """
  Represents a single spark event with its visual parameters.
  Uses an embedded Ecto schema and changesets for typing and validation.
  """
  use Ecto.Schema
  import Ecto.Changeset
  @derive Jason.Encoder

  @typedoc """
  A spark event to be broadcasted to the frontend.

  - `x`: The x-coordinate (0-511).
  - `y`: The y-coordinate (0-511).
  - `color`: A 3-element list representing HSL: `[hue, saturation, lightness]`.
  - `radius`: The spark's growth radius (integer > 0).
  - `transparency`: The spark's initial transparency (number 0.0-1.0).
  - `time_to_grow`: The animation duration in milliseconds (integer > 0).
  """
  @primary_key false
  embedded_schema do
    field :x, :integer
    field :y, :integer
    field :color, {:array, :float}
    field :radius, :float
    field :transparency, :float
    field :time_to_grow, :integer
  end

  @doc """
  Builds a changeset for creating a spark event.
  """
  def changeset(params) do
    %__MODULE__{}
    |> cast(params, [:x, :y, :color, :radius, :transparency, :time_to_grow])
    |> validate_required([:x, :y, :color, :radius, :transparency, :time_to_grow])
    |> validate_number(:x, greater_than_or_equal_to: 0, less_than_or_equal_to: 511)
    |> validate_number(:y, greater_than_or_equal_to: 0, less_than_or_equal_to: 511)
    |> validate_number(:radius, greater_than: 0)
    |> validate_number(:transparency, greater_than_or_equal_to: 0.0, less_than_or_equal_to: 1.0)
    |> validate_number(:time_to_grow, greater_than: 0)
    |> validate_color_format()
  end

  defp validate_color_format(changeset) do
    validate_change(changeset, :color, fn :color, color_list ->
      case color_list do
        [h, s, l]
        when (is_integer(h) or is_float(h)) and h >= 0 and h <= 360 and
             (is_integer(s) or is_float(s)) and s >= 0 and s <= 100 and
             (is_integer(l) or is_float(l)) and l >= 0 and l <= 100 ->
          []
        _ ->
          [color: {"is not a valid HSL list [0-360, 0-100, 0-100]", [validation: :invalid_hsl]}]
      end
    end)
  end
end
