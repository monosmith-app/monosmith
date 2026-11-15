from orders import total


def test_total() -> None:
    assert total([2, 3]) == 5
