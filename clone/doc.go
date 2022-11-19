//nolint:lll
/*
The clone package provides the ability to verify that the functions for cloning
structures with complex fields work correctly.

# The main goal

Structures containing fields with types such as slice, mapping and pointer must
be cloned with an explicit copying of the data to which they point. For example:

  type myStruct struct {items []string}
  func (ms *myStruct) Clone() *myStruct {
      rv := *ms
      // Allocating separate memory for items and copy original items
      rv.items = make([]string, len(ms.items))
      copy(rv.items, ms.items)
      return &rv
  }

Adding a new field to myStruct, it is possible to forget to add allocation and
copy operations corresponding to this field, thus leading to a situation where
the original structure and the clone share the same memory. Usually, this is
not expected behavior and causes difficult-to-catch bugs.

The clone package functionality ([StructVerifier.Verify]) reveals situations
where changes to the clone data cause changes to the original structure data,
which usually means that the cloning operation is incorrect.

See also [verification restrictions].

[verification restrictions]: https://pkg.go.dev/github.com/r-che/testing/clone/#hdr-Only_exported_fields_cloning_can_be_verified

# How to use

An example of the configuration structure with complex fields and the Clone
method to make a clone of the configuration:

  type Config struct {
      Int64param   int64
      Int64list    []int64
      StringList   []string
      MapVals      map[string]any
      // XXX The following fields are not exported and cannot be verified:
      test int64
      _Test int64
  }

  func NewConfig() *Config {
      return &Config{}
  }

  func (c *Config) Clone() *Config {
      // Make a simple copy of the original structure
      rv := *c

      // XXX Further, we need to clone all complex fields (slices, maps, etc...)

      rv.Int64list = make([]int64, len(c.Int64list))
      copy(rv.Int64list, c.Int64list)

      rv.StringList = make([]string, len(c.StringList))
      copy(rv.StringList, c.StringList)

      rv.MapVals = make(map[string]any, len(c.MapVals))
      for k, v := range c.MapVals {
          rv.MapVals[k] = v
      }

      return &rv
  }

Now, we can verify Config.Clone method using clone.StructVerifier:

  package main

  import (
      "fmt"
      "github.com/r-che/testing/clone"
  )

  func main() {
      sv := clone.NewStructVerifier(
          // Creator function
          func() any { return NewConfig() },
          // Cloner function
          func(x any) any {
              if c, ok := x.(*Config); ok {
                  return c.Clone()
              }
              panic(fmt.Sprintf("unsupported type: got - %T, want - *Config", x))
      })

      if err := sv.Verify(); err != nil {
          fmt.Printf("ERROR: %v\n", err)
      } else {
          fmt.Printf("Verification successful")
      }
  }

See more examples in the Examples section.

*/
package clone
