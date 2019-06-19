import React from "react";
import { Query } from "react-apollo";
import gql from "graphql-tag";
import "./UserList.css";
import { withStyles } from "@material-ui/core/styles";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Tooltip,
  Paper,
  TableSortLabel,
  Typography,
  TextField
} from "@material-ui/core";

const styles = theme => ({
  root: {
    maxWidth: 700,
    marginTop: theme.spacing.unit * 3,
    overflowX: "auto",
    margin: "auto"
  },
  table: {
    minWidth: 700
  },
  textField: {
    marginLeft: theme.spacing.unit,
    marginRight: theme.spacing.unit,
    minWidth: 300
  }
});

class UserList extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      order: "asc",
      orderBy: "name",
      page: 0,
      rowsPerPage: 10,
      usernameFilter: ""
    };
  }

  handleSortRequest = property => {
    const orderBy = property;
    let order = "desc";

    if (this.state.orderBy === property && this.state.order === "desc") {
      order = "asc";
    }

    this.setState({ order, orderBy });
  };

  getFilter = () => {
    return this.state.usernameFilter.length > 0
      ? { name_contains: this.state.usernameFilter }
      : {};
  };

  handleFilterChange = filterName => event => {
    const val = event.target.value;

    this.setState({
      [filterName]: val
    });
  };

  render() {
    const { order, orderBy } = this.state;
    const { classes } = this.props;
    return (
      <Paper className={classes.root}>
        <Typography variant="h2" gutterBottom>
          User List
        </Typography>
        <TextField
          id="search"
          label="User Name Contains"
          className={classes.textField}
          value={this.state.usernameFilter}
          onChange={this.handleFilterChange("usernameFilter")}
          margin="normal"
          variant="outlined"
          type="text"
          InputProps={{
            className: classes.input
          }}
        />

        <Query
          query={gql`
            query usersPaginateQuery(
              $first: Int
              $offset: Int
              $orderBy: [_UserOrdering]
              $filter: _UserFilter
            ) {
              User(
                first: $first
                offset: $offset
                orderBy: $orderBy
                filter: $filter
              ) {
                id
                name
                avgStars
                numReviews
              }
            }
          `}
          variables={{
            first: this.state.rowsPerPage,
            offset: this.state.rowsPerPage * this.state.page,
            orderBy: this.state.orderBy + "_" + this.state.order,
            filter: this.getFilter()
          }}
        >
          {({ loading, error, data }) => {
            if (loading) return <p>Loading...</p>;
            if (error) return <p>Error</p>;

            return (
              <Table className={this.props.classes.table}>
                <TableHead>
                  <TableRow>
                    <TableCell
                      key="name"
                      sortDirection={orderBy === "name" ? order : false}
                    >
                      <Tooltip
                        title="Sort"
                        placement="bottom-start"
                        enterDelay={300}
                      >
                        <TableSortLabel
                          active={orderBy === "name"}
                          direction={order}
                          onClick={() => this.handleSortRequest("name")}
                        >
                          Name
                        </TableSortLabel>
                      </Tooltip>
                    </TableCell>
                    <TableCell
                      key="avgStars"
                      sortDirection={orderBy === "avgStars" ? order : false}
                      numeric
                    >
                      <Tooltip
                        title="Sort"
                        placement="bottom-end"
                        enterDelay={300}
                      >
                        <TableSortLabel
                          active={orderBy === "avgStars"}
                          direction={order}
                          onClick={() => this.handleSortRequest("avgStars")}
                        >
                          Average Stars
                        </TableSortLabel>
                      </Tooltip>
                    </TableCell>
                    <TableCell
                      key="numReviews"
                      sortDirection={orderBy === "numReviews" ? order : false}
                      numeric
                    >
                      <Tooltip
                        title="Sort"
                        placement="bottom-start"
                        enterDelay={300}
                      >
                        <TableSortLabel
                          active={orderBy === "numReviews"}
                          direction={order}
                          onClick={() => this.handleSortRequest("numReviews")}
                        >
                          Number of Reviews
                        </TableSortLabel>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {data.User.map(n => {
                    return (
                      <TableRow key={n.id}>
                        <TableCell component="th" scope="row">
                          {n.name}
                        </TableCell>
                        <TableCell numeric>
                          {n.avgStars ? n.avgStars.toFixed(2) : "-"}
                        </TableCell>
                        <TableCell numeric>{n.numReviews}</TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            );
          }}
        </Query>
      </Paper>
    );
  }
}

export default withStyles(styles)(UserList);
