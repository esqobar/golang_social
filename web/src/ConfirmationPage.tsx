import { API_URL } from "./App.tsx";
import { useNavigate, useParams } from "react-router-dom";

const ConfirmationPage = () => {
    const { token = '' } = useParams();
    const navigate = useNavigate();

    const handleConfirm = async () => {
        const response = await fetch(`${API_URL}/users/activate/${token}`, {
            method: 'PUT',
        });

        if (response.ok) {
            navigate('/');
        } else {
            alert(`Failed to verify the confirm token ${response.status}`);
        }
    };

    return (
        <div>
            <h1>Confirmation</h1>
            <button onClick={handleConfirm}>Click to confirm</button>
        </div>
    );
};

export default ConfirmationPage;