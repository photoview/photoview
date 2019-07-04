import React, { Component } from "react";
import gql from 'graphql-tag'
import { Query } from 'react-apollo'

class App extends Component {
  render() {
    return (
      <div>
        <h1>Todo App</h1>
        <Query query={gql`
          query Todos {
            Todo {
              id
              title
            }
          }
        `}>
          {({data, loading, error}) => {
            if (loading) return <div>Loading todos...</div>
            if (error) return <div>Error</div>

            let todos = data.Todo.map(todo => <li key={todo.id}>{todo.title}</li>)

            return <ul>{todos}</ul>
          }}
        </Query>
      </div>
    );
  }
}

export default App;
