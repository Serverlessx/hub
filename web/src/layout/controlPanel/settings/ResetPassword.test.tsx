import { fireEvent, render, waitFor } from '@testing-library/react';
import React from 'react';
import { mocked } from 'ts-jest/utils';

import { API } from '../../../api';
import alertDispatcher from '../../../utils/alertDispatcher';
import ResetPassword from './ResetPassword';
jest.mock('../../../api');
jest.mock('../../../utils/alertDispatcher');

describe('Reset password - user settings', () => {
  afterEach(() => {
    jest.resetAllMocks();
  });

  it('creates snapshot', () => {
    const result = render(<ResetPassword />);

    expect(result.asFragment()).toMatchSnapshot();
  });

  describe('Render', () => {
    it('renders component', () => {
      const { getByTestId } = render(<ResetPassword />);

      const form = getByTestId('resetPasswordForm');
      expect(form).toBeInTheDocument();
      expect(getByTestId('oldPasswordInput')).toBeInTheDocument();
      expect(getByTestId('passwordInput')).toBeInTheDocument();
      expect(getByTestId('confirmPasswordInput')).toBeInTheDocument();
    });

    it('updates all fields and calls updatePassword', async () => {
      const { getByTestId } = render(<ResetPassword />);

      const oldPassword = getByTestId('oldPasswordInput') as HTMLInputElement;
      const newPassword = getByTestId('passwordInput') as HTMLInputElement;
      const repeatNewPassword = getByTestId('confirmPasswordInput') as HTMLInputElement;

      fireEvent.change(oldPassword, { target: { value: 'oldpass' } });
      fireEvent.change(newPassword, { target: { value: 'newpass' } });
      fireEvent.change(repeatNewPassword, { target: { value: 'newpass' } });

      const btn = getByTestId('updatePasswordBtn');
      expect(btn).toBeInTheDocument();
      fireEvent.click(btn);

      await waitFor(() => {});

      expect(API.updatePassword).toBeCalledTimes(1);
      expect(API.updatePassword).toHaveBeenCalledWith('oldpass', 'newpass');
    });

    it("doesn`t pass form validation when passwords don't match", async () => {
      const { getByTestId } = render(<ResetPassword />);

      const oldPassword = getByTestId('oldPasswordInput') as HTMLInputElement;
      const newPassword = getByTestId('passwordInput') as HTMLInputElement;
      const repeatNewPassword = getByTestId('confirmPasswordInput') as HTMLInputElement;

      fireEvent.change(oldPassword, { target: { value: 'oldpass' } });
      fireEvent.change(newPassword, { target: { value: 'new' } });
      fireEvent.change(repeatNewPassword, { target: { value: 'notMatch' } });

      const btn = getByTestId('updatePasswordBtn');
      expect(btn).toBeInTheDocument();
      fireEvent.click(btn);

      await waitFor(() => {});

      expect(API.updatePassword).toBeCalledTimes(0);
    });
  });

  describe('when updateUserProfile fails', () => {
    it('error 400', async () => {
      mocked(API).updatePassword.mockRejectedValue({
        status: 400,
        statusText: 'Error 400',
      });

      const { getByTestId } = render(<ResetPassword />);

      fireEvent.change(getByTestId('oldPasswordInput'), { target: { value: 'oldpass' } });
      fireEvent.change(getByTestId('passwordInput'), { target: { value: 'newpass' } });
      fireEvent.change(getByTestId('confirmPasswordInput'), { target: { value: 'newpass' } });

      const btn = getByTestId('updatePasswordBtn');
      expect(btn).toBeInTheDocument();
      fireEvent.click(btn);

      await waitFor(() => {
        expect(API.updatePassword).toHaveBeenCalledTimes(1);
      });

      expect(alertDispatcher.postAlert).toHaveBeenCalledTimes(1);
      expect(alertDispatcher.postAlert).toHaveBeenCalledWith({
        type: 'danger',
        message: 'An error occurred updating your password: Error 400',
      });
    });

    it('error 401', async () => {
      mocked(API).updatePassword.mockRejectedValue({
        statusText: 'ErrLoginRedirect',
      });

      const { getByTestId } = render(<ResetPassword />);

      fireEvent.change(getByTestId('oldPasswordInput'), { target: { value: 'oldpass' } });
      fireEvent.change(getByTestId('passwordInput'), { target: { value: 'newpass' } });
      fireEvent.change(getByTestId('confirmPasswordInput'), { target: { value: 'newpass' } });

      const btn = getByTestId('updatePasswordBtn');
      expect(btn).toBeInTheDocument();
      fireEvent.click(btn);

      await waitFor(() => {
        expect(API.updatePassword).toHaveBeenCalledTimes(1);
      });

      expect(alertDispatcher.postAlert).toHaveBeenCalledTimes(1);
      expect(alertDispatcher.postAlert).toHaveBeenCalledWith({
        type: 'danger',
        message:
          'An error occurred updating your password, please make sure you have entered your old password correctly',
      });
    });

    it('default error message', async () => {
      mocked(API).updatePassword.mockRejectedValue({
        status: 500,
      });

      const { getByTestId } = render(<ResetPassword />);

      fireEvent.change(getByTestId('oldPasswordInput'), { target: { value: 'oldpass' } });
      fireEvent.change(getByTestId('passwordInput'), { target: { value: 'newpass' } });
      fireEvent.change(getByTestId('confirmPasswordInput'), { target: { value: 'newpass' } });

      const btn = getByTestId('updatePasswordBtn');
      expect(btn).toBeInTheDocument();
      fireEvent.click(btn);

      await waitFor(() => {
        expect(API.updatePassword).toHaveBeenCalledTimes(1);
      });

      expect(alertDispatcher.postAlert).toHaveBeenCalledTimes(1);
      expect(alertDispatcher.postAlert).toHaveBeenCalledWith({
        type: 'danger',
        message: 'An error occurred updating your password, please try again later',
      });
    });
  });
});
