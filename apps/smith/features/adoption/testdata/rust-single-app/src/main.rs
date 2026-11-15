fn total(amounts: &[u32]) -> u32 {
    amounts.iter().sum()
}

fn main() {
    let mut buffer = itoa::Buffer::new();
    println!("{}", buffer.format(total(&[2, 3])));
}

#[cfg(test)]
mod tests {
    use super::*;
    use pretty_assertions::assert_eq;

    #[test]
    fn adds_the_amounts() {
        assert_eq!(total(&[2, 3]), 5);
    }
}
