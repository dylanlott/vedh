import { gql } from '@apollo/client/core';

export const CARD_QUERY = gql`
  query Card($id: String, $name: String!) {
    card(id: $id, name: $name) {
      ID
      Name
      Text
      CMC
      Types
    }
  }
`;

export const CARD_SEARCH_QUERY = gql`
  query SearchCards($name: String) {
    search(name: $name) {
      ID
      Name
      Types
    }
  }
`;
