input :: IO [Int]
input = do
    line <- getLine
    return (map read (words line))

main = do
    [a, b] <- input
    print (a + b)
