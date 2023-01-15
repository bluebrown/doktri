# Python Cheat Sheet for JavaScripters

Python has a lot in common with JavaScript. Functions are first class citizens and everything is an object.

For this post I will keep explanations short and try to let the code explain itself. I also expect the readers to be already familiar with most of the concepts, from lets say their JavaScript background. This post serves as mere cheat sheet for syntax.

## String Interpolation

```python
foo = 'bar'
baz = f'foo, {foo}.'
print(baz)

# Output: foo, bar.
```

## Ternary Operator

```python
hungry = True
action = 'eat' if hungry else 'relax'
print(f'Lets {action}')

# Output: Lets eat
```

## Short Circuit Evaluation

```python
thirsty = True

drink = {
  'type': 'banana split',
  'empty': True,
}

def makeDrink(drink_type):
  return {
  'type': drink_type,
  'empty': False,
}

drink_from = thirsty and (not drink['empty'] and drink or makeDrink('gin tonic'))

print(drink_from)

# Output: {'type': 'gin tonic', 'empty': False}
```

## Splat Operator / Unpacking - (Spread)

```python
eatables = {
    'mango': '1$',
    'banana': '0.5$'
}

drinks = {
    'water': '0.3$',
    'wine': {
        'white': '2$',
        'red': '1.5$'
    }
}

keys = {
    *eatables,
    *drinks
}

menu = {
    **eatables,
    **drinks
}

print(keys)
print(menu)

# Output: {'wine', 'mango', 'banana', 'water'}
# Output: {'mango': '1$', 'banana': '0.5$', 'water': '0.3$', 'wine': {'white': '2$', 'red': '1.5$'}}
```

## Anonymous Function

```python
double = lambda x: x * 2
print(double(2))

# Output: 4
```

## Immediate Invoked Function Expression

```python
(lambda name: print(name))('foo')

# Output: foo
```

## Callback

```python
def take_five(callback):
    callback(5)

take_five(lambda x: print(x*x))

# Output: 25
```

## Closure

```python
def outer_function(text):
    (lambda : # inner function
        print(text)
    )()

outer_function('bar')

# Output: bar
```

## Currying

```python
def multiplier_factory(factor):
    return lambda x: x*factor

doubler = multiplier_factory(2)
print(doubler(4))

# Output: 8
```

## Map

```python
items = [1, 2, 3, 4, 5]
map_obj = map(lambda x: x**2, items) # returns iterable object
print(list(map_obj)) # use list to iterate

# Output: [1, 4, 9, 16, 25]
```

## Filter

```python
negative_numbers = filter(lambda x: x < 0, range(-5, 5))
print(list(negative_numbers))

# Output: [-5, -4, -3, -2, -1]
```

## Zip

```python
a = ("John", "Charles", "Mike")
b = ("Jenny", "Christy", "Monica", "Vicky")
x = zip(a, b)
print(tuple(x)) # Read with tuple

# Output: (('John', 'Jenny'), ('Charles', 'Christy'), ('Mike', 'Monica'))
```

## List Comprehension

```python
numbers = [1, 2, 3, 4]
list_comp = [n**2 for n in numbers] # defined by square brackets
print(list_comp)

# Output: [1, 4, 9, 16]
```

### With Conditionals

```python
numbers = [1, 2, 3, 4]
list_comp = [n**2 for n in numbers if n%2 == 0 and n**n > 10]
print(list_comp)

# Output: [16]
```

### With Key-Value Pairs

```python
tuples = [('A','Good'), ('C','Faulty'), ('E','Pristine')]
something = [[key+' is '+val] for key, val in tuples if val != 'Faulty']
print(something)

# Output: [['A is Good'], ['E is Pristine']]
```

### With Index

```python
li = ['a', 'b', 'c']
op = [f'{x}_{i}' for i, x in enumerate(li)] # enum returns a list of tuples
print(op)

# Output: ['a_0', 'b_1', 'c_2']
```

### Indirect Nested

```python
words_list = [['foo', 'bar'], ['baz', 'impossibru!']]
merged = [word for words in words_list for word in words]
print(merged)

# Output: ['foo', 'bar', 'baz', 'impossibru!']
```

### Direct and Indirect Nested Madness

```python
split_merged = [letr for split_word in [[letter for letter in word] for words in words_list for word in words] for letr in split_word]
print(split_merged)

# Output: ['f', 'o', 'o', 'b', 'a', 'r', 'b', 'a', 'z', 'i', 'm', 'p', 'o', 's', 's', 'i', 'b', 'r', 'u', '!']
```

## Dict Comprehension

```python
rates = ['UYU_39.6', 'EUR_0.8']
op = {rate.split('_')[0]: rate.split('_')[1] for rate in rates} # defined by curly brackets
print(op)

# Output: {'UYU': '39.6', 'EUR': '0.8'}
```

## Generator Expression

```python
iterator = (item for item in ['a', 'b', 'c']) # defined by round brackets
result = next(iterator)
result = next(iterator)
iterator.close()
print(result)

# Output: b
```

## Async/Await

```python
import asyncio

async def simple():
    print("count one")
    await asyncio.sleep(1)
    print("count two")

await simple()

# Output: count one
# Output: count two
```

### With Gather

```python
import time

async def count_a():
    print("one")
    await asyncio.sleep(1)
    print("four")

async def count_b():
    print("two")
    await asyncio.sleep(1)
    print("five")

async def count_c():
    print("three")
    await asyncio.sleep(1)
    print("six")

async def gather_example():
    await asyncio.gather(
        count_a(),
        count_b(),
        count_c()
    )


s = time.perf_counter()

await gather_example()

elapsed = time.perf_counter() - s
print(f"Script executed in {elapsed:0.2f} seconds.")

# Output: one
# Output: two
# Output: three
# Output: four
# Output: five
# Output: six
# Output: Script executed in 1.00 seconds.
```
